import { toNextJsHandler } from "better-auth/next-js";
import { auth } from "@/lib/auth";

/**
 * Alle Anmelde-Endpunkte unter /api/auth/*.
 *
 * Eine Datei, zwei Methoden — Better Auth bringt das Routing selbst mit. Unter
 * dieser Adresse liegen unter anderem /api/auth/sign-in/email,
 * /api/auth/token und /api/auth/jwks.
 */
export const { GET, POST } = toNextJsHandler(auth);
