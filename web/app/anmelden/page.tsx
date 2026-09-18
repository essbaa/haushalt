import { Anmeldeformular } from "@/app/anmelden/formular";
import { mailErreichtAlle } from "@/lib/mail";

/**
 * Die Anmeldeseite.
 *
 * Eine Server Component über einer Client Component — und der einzige Grund
 * dafür ist eine Frage, die nur der Server beantworten kann: **Kann diese App
 * gerade Mail zustellen?** Der Bildschirm nach der Registrierung sagt je
 * nachdem etwas anderes, und was er sagt, muss stimmen.
 *
 * Solange die App keine eigene Domain hat, steht dort kein Satz über eine
 * Mail, die nie ankommt. Sobald `MAIL_VON` auf eine verifizierte Domain zeigt,
 * steht er wieder da — ohne dass hier jemand etwas umlegt.
 */
export default function Seite() {
  return <Anmeldeformular mailAn={mailErreichtAlle()} />;
}
