package storage

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zakaria/haushalt/api/internal/planner"
	"github.com/zakaria/haushalt/api/internal/storage/db"
)

// Gueltigkeit einer Einladung. Eine Woche ist lang genug, um sie in Ruhe
// weiterzugeben, und kurz genug, dass ein alter Code in einem Chatverlauf
// keinen Haushalt mehr öffnet.
const Gueltigkeit = 7 * 24 * time.Hour

// RoleOf ist die Rolle des Aufrufers in diesem Haushalt.
//
// Leer heißt: kein Mitglied. Das ist bei Demo-Haushalten der Normalfall und
// kein Fehler — sie sind öffentlich.
func (p *Plans) RoleOf(ctx context.Context, subject, id string) (planner.Role, error) {
	if subject == "" {
		return "", nil
	}
	zeile, err := p.lookup(ctx, id)
	if err != nil {
		return "", err
	}
	rolle, err := p.db.RoleInHousehold(ctx, db.RoleInHouseholdParams{
		HouseholdID: zeile.ID,
		AuthUserID:  &subject,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return planner.Role(rolle), nil
}

// Invite erzeugt einen einmalig gültigen Code.
//
// Nur planende Personen dürfen einladen, und die Rolle steht in der Einladung
// — nicht im Beitritt. Wer dazukommt, soll nicht selbst entscheiden, was er im
// Haushalt darf.
func (p *Plans) Invite(ctx context.Context, subject, id string, rolle planner.Role, mitgliedID string) (planner.Invitation, error) {
	if subject == "" {
		return planner.Invitation{}, planner.ErrNotAllowed
	}

	zeile, err := p.lookup(ctx, id)
	if err != nil {
		return planner.Invitation{}, err
	}

	// Erst Mitgliedschaft, dann Rolle. Ein Fremder bekommt „gibt es nicht",
	// ein Ausführender „nicht erlaubt" — der Unterschied ist Absicht.
	if err := p.mayAccess(ctx, subject, zeile); err != nil {
		return planner.Invitation{}, err
	}
	eigene, err := p.db.RoleInHousehold(ctx, db.RoleInHouseholdParams{
		HouseholdID: zeile.ID,
		AuthUserID:  &subject,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return planner.Invitation{}, fmt.Errorf("%w: %q", planner.ErrUnknownHousehold, id)
	}
	if err != nil {
		return planner.Invitation{}, err
	}
	if planner.Role(eigene) != planner.RolePlanner {
		return planner.Invitation{}, fmt.Errorf("%w: nur planende personen dürfen einladen", planner.ErrNotAllowed)
	}

	// Zeigt die Einladung auf jemanden, der schon im Plan steht, kommt die
	// Rolle von dieser Person — nicht aus der Anfrage. Sie hat ihre Rolle
	// bereits, der Plan rechnet damit, und eine Einladung ist kein Ort, an dem
	// man sie nebenbei ändert.
	var fuer pgtype.UUID
	var name string
	if mitgliedID != "" {
		kennung, ok := parseUUID(mitgliedID)
		if !ok {
			return planner.Invitation{}, fmt.Errorf("%w: %q ist keine person", planner.ErrNotAllowed, mitgliedID)
		}
		person, err := p.db.GetInvitableMember(ctx, db.GetInvitableMemberParams{
			ID:          kennung,
			HouseholdID: zeile.ID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			// Fremder Haushalt, betreute Person oder längst angemeldet — für
			// den Aufrufer ist das dieselbe Auskunft. Wer raten will, wer von
			// den dreien zutrifft, soll es nicht an der Antwort ablesen können.
			return planner.Invitation{}, fmt.Errorf("%w: diese person kann nicht eingeladen werden", planner.ErrNotAllowed)
		}
		if err != nil {
			return planner.Invitation{}, err
		}
		fuer = person.ID
		name = person.Name
		rolle = planner.Role(person.Role)
	}

	if rolle != planner.RolePlanner && rolle != planner.RoleDoer {
		// Wer betreut wird, meldet sich nicht an.
		return planner.Invitation{}, fmt.Errorf("%w: rolle %q kann nicht eingeladen werden", planner.ErrNotAllowed, rolle)
	}

	// Mit Haushalt gefragt, nicht nur mit der Anmeldung: Dieselbe Person kann
	// in mehreren Haushalten Mitglied sein, und created_by soll auf die
	// Person in DIESEM Haushalt zeigen.
	mitglied, err := p.db.GetMemberInHousehold(ctx, db.GetMemberInHouseholdParams{
		HouseholdID: zeile.ID,
		AuthUserID:  &subject,
	})
	if err != nil {
		return planner.Invitation{}, err
	}

	code, err := neuerCode()
	if err != nil {
		return planner.Invitation{}, err
	}
	bis := time.Now().Add(Gueltigkeit)
	if _, err := p.db.CreateInvitation(ctx, db.CreateInvitationParams{
		Code:        code,
		HouseholdID: zeile.ID,
		Role:        string(rolle),
		CreatedBy:   mitglied.ID,
		ExpiresAt:   zeitpunkt(bis),
		MemberID:    fuer,
	}); err != nil {
		return planner.Invitation{}, err
	}
	return planner.Invitation{Code: code, Until: bis, For: name, Role: rolle}, nil
}

// Accept verbindet die angemeldete Person mit dem Haushalt aus der Einladung.
//
// Ist sie schon Mitglied, passiert nichts weiter — der Code bleibt dann
// allerdings verbraucht, weil er benutzt wurde. Das ist die einfachere und
// ehrlichere Variante: Ein Code, der bei einem Fehlversuch gültig bleibt, ist
// ein Code, den man durchprobieren kann.
func (p *Plans) Accept(ctx context.Context, subject, name, code string) (planner.Household, error) {
	if subject == "" {
		return planner.Household{}, planner.ErrNotAllowed
	}

	einladung, err := p.db.GetInvitation(ctx, code)
	if errors.Is(err, pgx.ErrNoRows) {
		return planner.Household{}, planner.ErrUnknownInvitation
	}
	if err != nil {
		return planner.Household{}, err
	}
	switch {
	case einladung.UsedAt.Valid:
		return planner.Household{}, planner.ErrUnknownInvitation
	case !einladung.ExpiresAt.Valid || einladung.ExpiresAt.Time.Before(time.Now()):
		return planner.Household{}, planner.ErrUnknownInvitation
	}

	zeile, err := p.db.GetHousehold(ctx, einladung.HouseholdID)
	if err != nil {
		return planner.Household{}, err
	}

	// Schon dabei? Dann nur den Code verbrauchen.
	if _, err := p.db.RoleInHousehold(ctx, db.RoleInHouseholdParams{
		HouseholdID: zeile.ID,
		AuthUserID:  &subject,
	}); err == nil {
		if _, err := p.db.ConsumeInvitation(ctx, db.ConsumeInvitationParams{Code: code}); err != nil &&
			!errors.Is(err, pgx.ErrNoRows) {
			return planner.Household{}, err
		}
		return p.household(ctx, zeile)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return planner.Household{}, err
	}

	if name == "" {
		name = "Ich"
	}

	// Mitglied anlegen und Einladung verbrauchen gehören zusammen.
	//
	// Ohne Transaktion gibt es zwei Arten, es falsch zu machen: Das Mitglied
	// entsteht und die Markierung scheitert — dann gilt der Code ein zweites
	// Mal. Oder umgekehrt, und jemand steht ohne Haushalt da. Beides ist
	// selten und beides wäre schwer zu finden.
	tx, err := p.db.Pool.Begin(ctx)
	if err != nil {
		return planner.Household{}, err
	}
	// Nach einem erfolgreichen Commit ist das ein Nichtstun.
	defer func() { _ = tx.Rollback(ctx) }()

	q := p.db.Queries.WithTx(tx)

	// Zwei Wege, und der Unterschied ist der Grund für diese Migration: Zeigt
	// die Einladung auf eine Person, die schon im Plan steht, wird sie
	// übernommen. Sonst kommt jemand Neues dazu.
	//
	// Ohne den ersten Weg wurde aus „Asmae" beim Beitritt „Asmae (2)" — mit
	// halber Kapazität, ohne ihren Verlauf, und die Bilanz zeigte zwei
	// Menschen, wo einer sitzt.
	var mitglied db.Member
	if einladung.MemberID.Valid {
		mitglied, err = q.ClaimMember(ctx, db.ClaimMemberParams{
			AuthUserID:  &subject,
			ID:          einladung.MemberID,
			HouseholdID: zeile.ID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			// Jemand war schneller, oder die Person hat inzwischen ein Konto.
			// Für den Aufrufer ist das eine erschöpfte Einladung.
			return planner.Household{}, planner.ErrUnknownInvitation
		}
	} else {
		mitglied, err = q.CreateMemberWithRole(ctx, db.CreateMemberWithRoleParams{
			HouseholdID: zeile.ID,
			Name:        eindeutigerName(ctx, p, zeile, name),
			Role:        einladung.Role,
			// Geraten und änderbar. Ausführende bekommen weniger, weil hinter
			// der Rolle meist ein Kind oder Jugendlicher steckt.
			CapacityMinutes: kapazitaet(planner.Role(einladung.Role)),
			AuthUserID:      &subject,
		})
	}
	if err != nil {
		return planner.Household{}, err
	}

	// Erst jetzt der Anspruch auf die Einladung — mit den Bedingungen in der
	// Anweisung selbst. Kommt keine Zeile zurück, war jemand schneller, und
	// das Zurückrollen löscht das eben angelegte Mitglied wieder.
	if _, err := q.ConsumeInvitation(ctx, db.ConsumeInvitationParams{
		Code:   code,
		UsedBy: mitglied.ID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return planner.Household{}, planner.ErrUnknownInvitation
		}
		return planner.Household{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return planner.Household{}, err
	}
	return p.household(ctx, zeile)
}

// kapazitaet ist die Vorgabe für jemanden, der über eine Einladung dazukommt
// und nichts über seine Zeit gesagt hat.
//
// Die Zahlen stehen nicht hier, sondern in planner.TimeBudget: Sobald es zwei
// Stellen mit Minutenwerten gäbe, würden sie auseinanderlaufen.
func kapazitaet(rolle planner.Role) []int32 {
	if rolle == planner.RoleDoer {
		return minutenFelder(planner.BudgetLow)
	}
	return minutenFelder(planner.BudgetMedium)
}

// eindeutigerName weicht aus, wenn der Name im Haushalt schon vergeben ist —
// die Datenbank verlangt ihn eindeutig (siehe Migration 00002), und ein
// Beitritt soll nicht daran scheitern, dass zwei Menschen Anna heißen.
func eindeutigerName(ctx context.Context, p *Plans, z db.Household, name string) string {
	mitglieder, err := p.db.ListMembers(ctx, z.ID)
	if err != nil {
		return name
	}
	vergeben := map[string]bool{}
	for _, m := range mitglieder {
		vergeben[m.Name] = true
	}
	if !vergeben[name] {
		return name
	}
	for i := 2; i < 100; i++ {
		kandidat := fmt.Sprintf("%s (%d)", name, i)
		if !vergeben[kandidat] {
			return kandidat
		}
	}
	return name
}

// neuerCode erzeugt acht Zeichen aus einem Alphabet ohne Verwechslungsgefahr.
//
// Kein I, O, 0 oder 1: Der Code wird abgetippt oder durchtelefoniert. 32^8
// sind rund eine Billion Möglichkeiten — bei einmaliger Gültigkeit und einer
// Woche Haltbarkeit reicht das weit.
func neuerCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	roh := make([]byte, 8)
	if _, err := rand.Read(roh); err != nil {
		return "", err
	}
	out := make([]byte, len(roh))
	for i, b := range roh {
		out[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(out), nil
}
