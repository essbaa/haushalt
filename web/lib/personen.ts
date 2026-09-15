/**
 * Die Farbe einer Person.
 *
 * Jedes Mitglied bekommt einen festen Ton, und zwar nach seiner Stelle im
 * Haushalt. Damit ist die Zuteilung stabil, solange niemand geht — und über
 * alle Ansichten hinweg dieselbe: Wer im Plan blau ist, ist es in der Bilanz
 * auch.
 *
 * Nicht über eine Prüfsumme der Kennung, obwohl das ohne Liste auskäme: Sie
 * verteilt zufällig und legt in einem Haushalt mit drei Personen gern zwei
 * davon auf denselben Ton. Bei sechs Farben und drei Personen ist das kein
 * seltener Unfall, sondern jede dritte Familie.
 *
 * Die Farbe sagt nie etwas allein — daneben steht immer der Name (WCAG 1.4.1).
 * Sie ist eine Abkürzung beim Lesen, keine Auskunft für sich.
 */

const toene = 6;

/** Die CSS-Klasse, die --person und --person-soft setzt. */
export function farbklasse(mitglieder: { id: string }[], id: string): string {
  if (id === "") return "";
  const stelle = mitglieder.findIndex((m) => m.id === id);
  if (stelle < 0) return "";
  return `person-${(stelle % toene) + 1}`;
}
