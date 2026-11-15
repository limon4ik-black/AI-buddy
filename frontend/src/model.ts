import type {
  Account,
  Profile,
  Checklist,
  Chat,
  Team,
  TeamDetail,
  Service,
  Document,
  Tool,
  AccessRequest,
} from "./api";
export type View =
  | "home"
  | "checklist"
  | "chat"
  | "profile"
  | "teams"
  | "services"
  | "knowledge"
  | "notes"
  | "boards";
export type Workspace = {
  account: Account;
  profile: Profile;
  checklist: Checklist;
  chats: Chat[];
  teams: Team[];
  teamDetails: Record<string, TeamDetail>;
  services: Service[];
  documents: Document[];
  tools: Tool[];
  requests: AccessRequest[];
};
export const teamColors: Record<string, string> = {
  backend: "blue",
  frontend: "purple",
  analytics: "green",
  qa: "red",
};
export const teamNames: Record<string, string> = {
  backend: "Backend",
  frontend: "Frontend",
  analytics: "Analytics",
  qa: "QA",
};
export const demoAccounts = [
  { email: "backend@buddy.test", name: "Егор", teamId: "backend" },
  { email: "frontend@buddy.test", name: "Анна", teamId: "frontend" },
  { email: "analytics@buddy.test", name: "Максим", teamId: "analytics" },
  { email: "qa@buddy.test", name: "София", teamId: "qa" },
  { email: "backend2@buddy.test", name: "Алина", teamId: "backend" },
];
export const demoPassword = "BuddyDemo2026!";
export const initials = (name: string) =>
  name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((w) => w[0])
    .join("")
    .toUpperCase();
export const dateLabel = (date: string) =>
  new Intl.DateTimeFormat("ru-RU", { day: "numeric", month: "long" }).format(
    new Date(date),
  );
export const roleLabel: Record<string, string> = {
  lead: "Тимлид",
  buddy: "Бадди",
  engineer: "Инженер",
  intern: "Стажёр",
};
export const safeExternalUrl = (value: string) => {
  try {
    const url = new URL(value);
    return ["http:", "https:"].includes(url.protocol) ? url.href : undefined;
  } catch {
    return undefined;
  }
};
