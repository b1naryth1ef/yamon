export type YamonLogEntry = {
  service: string;
  level: string;
  data: any;
  time?: number;
  tags?: Record<string, string>;
};

export type YamonEvent = {
  type: string;
  data: any;
  time?: number;
  tags?: Record<string, string>;
};

export type YamonMetric = {
  type: "counter" | "gauge";
  name: string;
  value: number;
  time?: number;
  tags?: Record<string, string>;
};

export type YamonScriptResult = {
  metrics?: Array<YamonMetric>;
  metric?: YamonMetric;
  logs?: Array<YamonLogEntry>;
  log?: YamonLogEntry;
  events?: Array<YamonEvent>;
  event?: YamonEvent;
};

export async function writeResult(result: YamonScriptResult) {
  const data = JSON.stringify(result);
  await Deno.stdout.write(new TextEncoder().encode(data + "\n"));
}
