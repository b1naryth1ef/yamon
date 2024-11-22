export type YamonMetric = {
  type: "counter" | "gauge";
  name: string;
  value: number;
  time?: number;
  tags?: Record<string, string>;
};

export type YamonScriptResult = {
  metrics: Array<YamonMetric>;
};

export async function writeResult(result: YamonScriptResult) {
  const data = JSON.stringify(result);
  await Deno.stdout.write(new TextEncoder().encode(data));
}
