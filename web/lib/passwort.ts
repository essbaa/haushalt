/**
 * Wie stark ist dieses Passwort?
 *
 * Die naheliegende Rechnung zählt Zeichenklassen — groß, klein, Ziffer,
 * Sonderzeichen — und ist falsch herum. Sie gibt `Passwort1!` die Bestnote
 * (acht Zeichen, vier Klassen) und `pferdeklammerbatterie` die schlechteste,
 * obwohl das zweite um Größenordnungen schwerer zu raten ist. Wer rät, rät
 * nicht zeichenweise, sondern mit Wörterbüchern und den Mustern, die Menschen
 * benutzen: ein großer Anfangsbuchstabe, eine Ziffer am Ende, ein Ausrufezeichen
 * dahinter.
 *
 * Deshalb: **Länge gibt den Ausschlag, Klassen geben einen Bonus.** Die
 * Klassen werden trotzdem einzeln zurückgegeben, weil die Oberfläche sie
 * anzeigt — als Rat, nicht als Pflicht. Erzwungen erzeugen sie genau das
 * `Passwort1!`, gegen das diese Datei antritt (siehe ADR-0014).
 *
 * Kein zxcvbn: Die Bibliothek rechnet besser, wiegt aber mehrere hundert
 * Kilobyte im Bundle einer App, die auf dem Telefon geöffnet wird. Für die
 * Frage „ist das offensichtlich zu wenig" reicht das hier.
 */

export type Klassen = {
  klein: boolean;
  gross: boolean;
  ziffer: boolean;
  sonder: boolean;
};

export type Staerke = {
  /** 0 bis 4. 0 heißt: zu kurz, um überhaupt gewertet zu werden. */
  stufe: 0 | 1 | 2 | 3 | 4;
  text: string;
  klassen: Klassen;
  /** Gesetzt, wenn etwas Konkretes dagegen spricht — wiegt schwerer als die Stufe. */
  einwand?: string;
};

/** Die zwanzig, die in jeder Liste geleakter Passwörter oben stehen. */
const haeufig = new Set([
  "passwort",
  "password",
  "12345678",
  "123456789",
  "1234567890",
  "qwertz123",
  "qwerty123",
  "passwort1",
  "password1",
  "iloveyou",
  "sonnenschein",
  "hallo123",
  "willkommen",
  "welcome1",
  "admin123",
  "letmein1",
  "monkey123",
  "fussball",
  "dragon123",
  "abc12345",
]);

export function staerke(passwort: string, umfeld: string[] = []): Staerke {
  const klassen: Klassen = {
    klein: /[a-zäöüß]/.test(passwort),
    gross: /[A-ZÄÖÜ]/.test(passwort),
    ziffer: /[0-9]/.test(passwort),
    sonder: /[^A-Za-zÄÖÜäöüß0-9]/.test(passwort),
  };

  if (passwort.length === 0) {
    return { stufe: 0, text: "", klassen };
  }
  if (passwort.length < 8) {
    return { stufe: 0, text: "zu kurz", klassen, einwand: "Mindestens acht Zeichen." };
  }

  const klein = passwort.toLowerCase();

  if (haeufig.has(klein)) {
    return {
      stufe: 0,
      text: "sehr schwach",
      klassen,
      einwand: "Dieses Passwort steht in jeder Liste, mit der geraten wird.",
    };
  }

  // Der eigene Name oder die eigene Adresse darin ist der erste Versuch, den
  // jemand macht, der dich kennt.
  for (const wort of umfeld) {
    const w = wort.trim().toLowerCase();
    if (w.length >= 3 && klein.includes(w)) {
      return {
        stufe: 1,
        text: "schwach",
        klassen,
        einwand: "Enthält deinen Namen oder deine Adresse.",
      };
    }
  }

  // Ein einziges wiederholtes Zeichen oder eine Tastaturreihe ist lang und
  // trotzdem nichts wert.
  if (
    /^(.)\1+$/.test(passwort) ||
    "abcdefghijklmnopqrstuvwxyz".includes(klein) ||
    "0123456789".includes(klein)
  ) {
    return { stufe: 1, text: "schwach", klassen, einwand: "Zu gleichförmig." };
  }

  let punkte = 1;
  if (passwort.length >= 12) punkte++;
  if (passwort.length >= 16) punkte++;

  const anzahl = Object.values(klassen).filter(Boolean).length;
  if (anzahl >= 3) punkte++;

  // Kurz UND einförmig: Ein Bonus für Länge, den es nicht gab, wird hier
  // wieder abgezogen.
  if (passwort.length < 12 && anzahl === 1) punkte--;

  const stufe = Math.max(1, Math.min(4, punkte)) as 1 | 2 | 3 | 4;
  return { stufe, text: ["", "schwach", "geht so", "stark", "sehr stark"][stufe], klassen };
}
