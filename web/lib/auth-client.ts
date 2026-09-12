"use client";

import { createAuthClient } from "better-auth/react";
import { jwtClient } from "better-auth/client/plugins";

/**
 * Der Anmelde-Client für den Browser.
 *
 * Ohne baseURL: Die Endpunkte liegen unter derselben Adresse wie die Seite.
 * Das ist der Unterschied zum Go-Dienst — der hat eine eigene Adresse und
 * deshalb CORS, die Anmeldung hat beides nicht.
 */
export const authClient = createAuthClient({
  plugins: [jwtClient()],
});

export const { signIn, signUp, signOut, useSession } = authClient;
