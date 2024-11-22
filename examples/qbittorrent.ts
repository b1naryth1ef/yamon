#!/usr/bin/env -S deno run -A --node-modules-dir=auto

const QBITTORRENT_HOST = Deno.env.get("QBITTORRENT_HOST");

type YamonMetric = {
  type: "counter" | "gauge";
  name: string;
  value: number;
  time?: number;
  tags?: Record<string, string>;
};

type YamonScriptResult = {
  metrics: Array<YamonMetric>;
};

async function writeResult(result: YamonScriptResult) {
  const data = JSON.stringify(result);
  await Deno.stdout.write(new TextEncoder().encode(data));
}

export async function main() {
  if (!QBITTORRENT_HOST) {
    throw new Error("Please set the QBITTORRENT_HOST env variable!");
  }

  const res = await fetch(`http://${QBITTORRENT_HOST}/api/v2/sync/maindata`);
  if (!res.ok) {
    throw new Error(
      `failed to fetch qbittorrent metadata: ${await res.text()}`
    );
  }

  const result: YamonScriptResult = { metrics: [] };

  const data = await res.json();
  result.metrics.push({
    type: "counter",
    name: "qbittorrent.server.alltime_dl",
    value: data.server_state.alltime_dl,
  });
  result.metrics.push({
    type: "counter",
    name: "qbittorrent.server.alltime_ul",
    value: data.server_state.alltime_ul,
  });
  result.metrics.push({
    type: "gauge",
    name: "qbittorrent.server.average_time_queue",
    value: data.server_state.average_time_queue,
  });
  result.metrics.push({
    type: "gauge",
    name: "qbittorrent.server.dht_nodes",
    value: data.server_state.dht_nodes,
  });
  result.metrics.push({
    type: "gauge",
    name: "qbittorrent.server.global_ratio",
    value: parseFloat(data.server_state.global_ratio),
  });

  return await writeResult(result);
}

await main();
