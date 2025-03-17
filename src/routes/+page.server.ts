import { getShifts } from "$lib/pb";
import type { ServerLoad } from "@sveltejs/kit";

export const load: ServerLoad = async ({ locals }) => {
  const shifts = await getShifts();
  return {
    shifts,
  };
};
