#!/usr/bin/env -S deno run -A --node-modules-dir=auto

const STREAMING = Deno.env.get("STREAMING");
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
  await Deno.stdout.write(new TextEncoder().encode(data + "\n"));
}

async function collect() {
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

  await writeResult(result);
}

export async function main() {
  if (!QBITTORRENT_HOST) {
    throw new Error("Please set the QBITTORRENT_HOST env variable!");
  }

  if (STREAMING === undefined) {
    await collect();
    return;
  } else {
    const loopFn = () => {
      collect().then(() => {
        setTimeout(loopFn, 1000);
      });
    };
    loopFn();
  }

  const p = new Promise(() => {});
  await p;
}

await main();
