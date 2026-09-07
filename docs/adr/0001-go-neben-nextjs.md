# 0001 — Go-Dienst neben Next.js statt Next.js allein

**Status:** angenommen · 7. September 2026

## Kontext

Next.js kann Backend. Route Handler, Server Actions, ein Datenbanktreiber im
selben Prozess — die gesamte Anwendung liefe als ein Dienst auf Vercel, mit
einer Sprache, einer Pipeline und ohne Netzwerkgrenze in der Mitte. Für die
meisten Anwendungen dieser Größe ist das die richtige Antwort, und jede
Abweichung davon braucht eine Begründung.

Vier Dinge sprechen hier dagegen:

**Der Kern ist ein Algorithmus, keine Datenschicht.** Der Wochenplaner wählt
Vorlagen aus, entscheidet Fälligkeit, begrenzt die Startdichte und verteilt
Aufgaben über zwei Achsen (Dauer und Kopflast). Das ist Logik mit echtem
Testbedarf — tabellengetriebene Tests über Dutzende Haushaltszuschnitte, ohne
HTTP, ohne Datenbank, in Millisekunden. Genau dafür ist ein reines Paket in
einer statisch typisierten Sprache gebaut.

**Kommende Arbeit hängt nicht am Browser.** Wochenpläne entstehen sonntags,
nicht wenn jemand die Seite öffnet. Erinnerungen laufen zeitgesteuert.
LLM-Aufrufe für Erklärtexte und Vorlagenvorschläge dauern Sekunden bis
Minuten. Alles davon will einen Prozess, der ohne eingehenden Request lebt und
keine Ausführungsgrenze einer Function-Laufzeit kennt.

**Betriebskosten sind kalkulierbar.** Eine kleine Maschine in Frankfurt mit
festem Preis ist bei langlaufenden Aufrufen berechenbarer als Abrechnung nach
Ausführungszeit.

**Ausdrücklich genannt: Go vertiefen ist Nebenziel.** Dieses Projekt ist
Produkt und Arbeitsprobe zugleich. Ein zweiter Dienst in einer Sprache, die
ich beherrschen will, ist Teil des Zwecks — nicht der Hauptgrund, aber ein
echter, und ein verschwiegener Grund wäre ein schlechter Eintrag.

## Entscheidung

Zwei Dienste in einem Repo: `web/` (Next.js, rendert und spricht ausschließlich
über HTTP mit dem Backend) und `api/` (Go, `net/http`, Standardbibliothek).
Der Planer liegt als reines Paket `api/internal/planner` ohne jedes I/O —
Eingabe rein, Ergebnis raus, keine Datenbank, keine Uhr, keine Zufallszahlen.

Kein Framework im Go-Dienst. `net/http` mit dem `ServeMux` ab Go 1.22 deckt
Methoden und Pfadparameter ab; für diesen Umfang trägt eine zusätzliche
Abhängigkeit nichts.

## Konsequenzen

**Dafür:**

- Der Kern ist ohne Infrastruktur testbar. Ein Planungslauf im Test kostet
  keine Sekunde und keine Datenbank.
- Die Sprachgrenze erzwingt einen ausformulierten Vertrag zwischen Anzeige und
  Logik. Was in einem Monolithen als Abkürzung durchginge, fällt hier auf.
- Beide Seiten werden getrennt ausgeliefert. Eine Änderung am Frontend kann den
  Planer nicht kaputt machen und umgekehrt; die Pfadfilter in der CI machen
  das sichtbar.
- Langlaufende Arbeit hat später einen Ort, an den sie gehört.

**Dagegen — der Preis, den wir zahlen:**

- **CORS.** Der erste echte Fehler zwischen den Diensten, und der, an dem man
  am längsten sitzt. Ist bereits eingetreten.
- **Zwei Auslieferungswege**, zwei Sätze Geheimnisse, zwei Laufzeiten, die
  aktuell gehalten werden wollen.
- **Typen stehen doppelt** — einmal als Go-Struktur, einmal als TypeScript-Typ.
  Bis der generierte Client aus einer OpenAPI-Beschreibung steht (T4), ist das
  Handarbeit und damit eine Fehlerquelle.
- **Lokal laufen zwei Prozesse.** Der Einstieg für einen zweiten Entwickler ist
  länger als ein `npm run dev`.
- **Ein Netzwerkaufruf mehr** zwischen Anzeige und Daten, samt Kaltstart der
  Maschine, wenn sie geruht hat.

## Verworfene Alternativen

**Next.js allein, Logik in Route Handlern.** Weniger Betrieb, keine
Sprachgrenze — aber der Planer läge in derselben Laufzeit wie das Rendern, die
Tests bräuchten die Next-Umgebung, und Hintergrundarbeit müsste über einen
externen Auslöser nachgerüstet werden. Der Preis wäre später fällig, nicht
heute.

**Go allein, Server-Templates oder HTMX.** Ein Dienst, kein CORS, kein
doppelter Typ. Verworfen aus Produktgründen: Die Wochenansicht mit Verschieben,
Tauschen und sofortiger Rückmeldung ist der Kern der Bedienung, und dafür ist
React das passendere Werkzeug — abgesehen davon, dass es meine Stärke ist.

**Node- oder Nest-Backend als eigener Dienst.** Hätte dieselbe Trennung
gebracht, dazu geteilte Typen ohne Generator. Verworfen wegen des Nebenziels
Go und weil ein reines, abhängigkeitsfreies Kernpaket in Go schlicht besser
aufgehoben ist.

## Wann wir das revidieren

Wenn nach drei Monaten alle drei Punkte zutreffen — der Planer rechnet nie
länger als eine Sekunde, es ist keine zeitgesteuerte Hintergrundarbeit
entstanden, und die LLM-Aufrufe laufen ohnehin über einen fremden Dienst mit
eigener Warteschlange — dann trägt der zweite Dienst seine Kosten nicht mehr.
In dem Fall wäre der Weg zurück offen: Der Planer ist ein reines Paket, seine
Portierung wäre Übersetzungsarbeit ohne Architekturfrage.
