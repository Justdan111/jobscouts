export type Eligibility = "global" | "us-only" | "unknown";
export type TriageStatus = "new" | "saved" | "dismissed";
export type AppStage = "drafted" | "applied" | "interviewing" | "offer" | "rejected";

export type User = { id: string; email: string };
export type Link = { label: string; url: string };

export type Profile = {
  userId: string;
  name: string;
  headline: string;
  basedIn: string;
  timezone: string;
  links: Link[];
  resumeText: string;
  resumeFile: string;
  roles: string[];
  stack: string[];
  seniorityWant: string[];
  jobTypes: string[];
  remoteOnly: boolean;
  eligibleNote: string;
  lookingFor: string;
  strengths: string;
  targets: string;
  constraints: string;
  extraContext: string;
  updatedAt: string;
};

export type Job = {
  id: string;
  source: string;
  title: string;
  company: string;
  location: string;
  url: string;
  description: string;
  tags: string[] | null;
  postedAt: string;
  eligibility: Eligibility;
  yc: boolean;
  ycBatch: string;
  funded: boolean;
  newlyFunded: boolean;
  internship: boolean;
  score: number;
  reason: string;
  status: TriageStatus;
  discoveredAt: string;
};

export type Company = {
  name: string;
  slug: string;
  batch: string;
  oneLiner: string;
  longDesc: string;
  website: string;
  ycUrl: string;
  location: string;
  industries: string[] | null;
  isHiring: boolean;
  remote: boolean;
};

export type Application = {
  userId: string;
  postingId: string;
  stage: AppStage;
  resumeHighlights: string;
  coverEmail: string;
  pitch: string;
  notes: string;
  createdAt: string;
  updatedAt: string;
};
