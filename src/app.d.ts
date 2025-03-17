// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
import PocketBase, { type AuthRecord } from "pocketbase";

declare global {
  namespace App {
    // interface Error {}
    interface Locals {
      pb: PocketBase;
      user: AuthRecord | null;
    }
    // interface PageData {}
    // interface PageState {}
    // interface Platform {}
  }
}

export {};
