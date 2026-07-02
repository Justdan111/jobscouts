import type { Job } from "./types";

// Backend now classifies jobs (yc / funded / newlyFunded / internship) using
// real YC data. These helpers just read those fields, with a light text
// fallback for the funded flag so older records still count.
const FALLBACK = ["(yc", "y combinator", "seed", "series a", "series b", "raised", "venture"];

export function isFunded(job: Job): boolean {
  if (job.funded) return true;
  const blob = `${job.company} ${(job.tags || []).join(" ")} ${job.description}`.toLowerCase();
  return FALLBACK.some((s) => blob.includes(s));
}
