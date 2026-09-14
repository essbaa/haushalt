# 0010 — Unbekannt ist nicht nein

**Status:** angenommen · 13. September 2026 · Nachtrag 14. September

## Kontext

Zakaria richtete seinen Haushalt ein und bekam einen Plan mit zwei Aufgaben,
die ihn nichts angingen:

- **„Pflanzen gießen"** — sie haben keine Pflanzen.
- **„Geschenk für Kindergeburtstag besorgen", Montag** — es steht kein
  Geburtstag an.

Beides waren keine Ausrutscher, sondern zwei verschiedene Konstruktionsfehler.

**Erstens kannte das Modell nur zwei Antworten.** Der Haushaltskontext hielt
Wohnform, Auto, Garten und Haustiere; Bedingungen gab es für Kinderalter und
Betreuung. Alles darüber hinaus — Pflanzen, Spülmaschine, Keller, Fahrrad,
Dunstabzug — kam schlicht nicht vor. Eine Vorlage ohne Bedingung galt für
jeden. Es fehlte nicht *eine* Bedingung, sondern die ganze Klasse „Dinge, die
es hier nicht gibt".

**Zweitens gab sich eine Periode als Anlass aus.** `t-geschenk-kindergeburtstag`
trug `rhythmus: { typ: "ausloeser", alle_tage: 45 }`. Es gibt keinen
Geburtstag in der App — die Vorlage erfand alle anderthalb Monate einen.

Das Produktkonzept sagt: *Der erste Plan ist der Moment, in dem die App
beweisen muss, dass sie etwas weiß, das man selbst vergessen hätte.* Ein Plan
mit erfundenen Pflanzen beweist das Gegenteil, und nach dem dritten falschen
Vorschlag glaubt niemand mehr den richtigen.

## Entscheidung

**Unbekannt ist ein eigener Zustand.** `Context.Facts` ist eine Karte
Name → Wahrheitswert. Ein fehlender Schlüssel heißt *unbekannt*, nicht *nein*.
Die Bedingungsprüfung hat vier Ausgänge statt zwei: gilt, gilt nicht, unklar,
braucht einen Termin.

**Was unklar ist, wird nicht eingeplant, sondern gefragt.** Aus jedem
unbekannten Faktum wird ein Fragenkandidat. Die App stellt **höchstens zwei
Fragen** je Woche, sortiert danach, an wie vielen Vorlagen das Faktum hängt.

**Jede Frage nennt ihren Nutzen.** „Habt ihr Pflanzen, die gegossen werden
müssen? — Dann erinnere ich alle paar Tage ans Gießen." Der Loader weist eine
Frage ohne Nutzen zurück; das ist bewusst hart, denn wer nicht weiß, wofür er
antwortet, antwortet nicht.

**Ohne Anlass keine Aufgabe.** `benoetigt_termin: true` heißt: Diese Aufgabe
entsteht aus einem Ereignis, nie aus einem Zeitraum. Bis es Termine in der App
gibt, steht sie in „Nicht im Plan" mit genau diesem Grund.

**Manche Vorlagen sind Angebote, keine Annahmen.** „Zeit zu zweit einplanen",
„Verabredung mit Freunden", „Bei den Eltern melden" hängen jetzt an einem
Faktum, das nur eine ausdrückliche Antwort setzt. Sie ungefragt einzuplanen
wäre übergriffig — und bei jemandem, dessen Eltern nicht mehr leben, schlimmer
als das.

**Kopflast wird auch je Tag begrenzt** (`MaxHeadLoadPerDay`, Vorgabe 4). Die
Startdichte zählt Organisationsaufgaben je Woche und sagt nichts darüber, wie
sie liegen — vier Terminsachen landeten alle am Montag. Die Grenze ist eine
Vorliebe, keine Bedingung: Erst sucht jede Aufgabe einen ruhigen Tag, findet
sie keinen, nimmt sie einen vollen.

## Konsequenzen

**Positiv.** Der erste Plan behauptet nichts mehr. 37 von 62 Vorlagen haben
jetzt eine Voraussetzung. Die Fragen ersetzen einen Fragebogen durch etwas, das
beantwortet wird, weil der Nutzen danebensteht. Und nebenbei wurde der Montag
leichter — als die Kopfarbeit sich verteilte, sank auch die Zahl der Aufgaben,
die aus Kapazitätsmangel wegfielen, von drei auf eine.

**Negativ.** Der erste Plan eines neuen Haushalts ist deutlich kleiner, und
manches Nützliche fehlt darin, bis jemand geantwortet hat. Das ist der Preis
für Ehrlichkeit, aber es ist ein Preis: Weniger Vorschläge heißt weniger
Gelegenheit zu beweisen, dass die App etwas weiß.

Die Fragen sind nur für planende Personen sichtbar. Wer ausführt, sieht einen
kleineren Plan und kann nichts daran ändern.

Und der Fragenkatalog ist Handarbeit. Jede neue Vorlage mit einer
Voraussetzung braucht ein Faktum mit einer Frage, sonst bleibt sie für immer
unsichtbar — der Loader prüft, dass Fragen einen Nutzen haben, aber nicht, dass
zu jedem Faktum eine Frage existiert.

## Verworfene Alternativen

**Im Onboarding abfragen.** Ein vierter Schritt mit fünfzehn Schaltern wäre am
einfachsten zu bauen und verletzt genau die Regel, an der laut Produktkonzept
die Konkurrenz scheitert: eine App gegen Mental Load, die beim Einrichten
welche erzeugt.

**Nur die Bibliothek kuratieren, ohne Fragen.** Ehrlich und endgültig: Die
weggelassenen Aufgaben kämen nie wieder, und die App lernte nichts über den
Haushalt.

**Unbekannt wie „ja" behandeln, und der Nutzer löscht.** Das war der Zustand,
den wir gerade abgeschafft haben. Er verschiebt die Arbeit auf den Nutzer und
kostet Vertrauen, bevor er Arbeit spart.

## Wann wir das revidieren

Sobald es Termine in der App gibt, bekommen die Vorlagen mit
`benoetigt_termin` ihren Anlass und kehren in den Plan zurück. Bis dahin sind
sie sichtbar abwesend, und das ist besser als sichtbar falsch.

Wenn sich zeigt, dass die Fragen liegen bleiben, ist nicht die Zahl zu klein,
sondern der Nutzen zu schwach formuliert — dann gehört an den Texten gearbeitet
und nicht am Mechanismus.


## Nachtrag, 14. September 2026

**Die Anlässe gibt es jetzt.** Unter „Wann wir das revidieren" stand, dass die
Vorlagen mit `benoetigt_termin` zurückkehren, sobald es Termine gibt — das ist
am selben Tag passiert. `occasion` hält Geburtstage, Elternabende und Termine;
Geburtstage wiederholen sich jährlich, und die Wiederholung hängt an der Art
statt an einem eigenen Schalter, weil die App die Antwort schon kennt.

„Geschenk für Kindergeburtstag besorgen" entsteht damit sieben Tage vor einem
eingetragenen Geburtstag, mit dem Geburtstag als Frist — ein Geschenk nach dem
Geburtstag ist kein Geschenk. Ohne eingetragenen Anlass entsteht nichts, und
genau das war der Punkt.

**Dazu eine zweite Lehre aus derselben Ecke:** Eine Frage, die nichts bewirkt,
ist derselbe Fehler wie eine Behauptung ohne Grundlage — nur andersherum. Der
Fakten-Loader wies von Anfang an eine Frage ohne erkennbaren Nutzen zurück, für
das Onboarding galt die Regel nicht, und die Wohnform war der Beweis. Sie wird
jetzt **mechanisch geprüft**: Ein Test in `internal/library` hält die Bibliothek
gegen sich selbst — jedes Faktum wird von mindestens einer Vorlage gebraucht,
jede Voraussetzung hat eine Frage, jede Anlassart lässt sich eintragen. Es ist
der erste Test im Projekt, der eine Produktregel prüft statt Code, und damit die
einzige Sorte, die die Fehler dieses Tages gefangen hätte.
