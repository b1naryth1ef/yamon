import { getCommit, task } from "@maf/core.ts";
import { Image, run } from "@maf/docker/mod.ts";
import { getGoBuildEnv, GOARCH, GoBuild, GOOS } from "@maf/lang/go.ts";
import {
  CommitStatus,
  getClient,
  Release,
  webhook,
} from "@maf/service/github.ts";
import { format as formatBytes } from "@std/fmt/bytes.ts";
import { upload } from "https://gist.githubusercontent.com/b1naryth1ef/4b649811c903e6c26283110167620cb4/raw/d579a8bd668db207eba2f17784c314f412803340/upload-artifact.ts";

const dockerFile = `
FROM golang:1.24-bullseye
RUN apt-get update -y && apt-get install -y build-essential libsystemd-dev:amd64 gcc-9-aarch64-linux-gnu
`;

export type Project = "agent" | "server";
const matrix: Array<[Project, GoBuild]> = [
  ["server", { os: GOOS.linux, arch: GOARCH.amd64 }],
  ["agent", { os: GOOS.linux, arch: GOARCH.amd64 }],
  ["agent", { os: GOOS.linux, arch: GOARCH.arm64 }],
];

export const build = task(
  "build",
  async ({
    go,
    project,
    release,
    latest,
  }: {
    go?: GoBuild;
    project: Project;
    release?: Release;
    latest?: boolean;
  }) => {
    go = go || { os: GOOS.linux, arch: GOARCH.amd64 };

    const name = `yamon-${project}-${go.os}-${go.arch}`;

    let commitStatus = null;
    const client = await getClient();
    if (client) {
      commitStatus = await CommitStatus.create(
        "b1naryth1ef/yamon",
        getCommit().id,
        {
          state: "pending",
          context: name,
        }
      );
    }

    const imageId = await Image.fromString(dockerFile);

    const env = [...getGoBuildEnv(go), "CGO_ENABLED=1"];
    if (go.arch === GOARCH.arm64) {
      env.push("CC=aarch64-linux-gnu-gcc-9");
    }

    await run(`go build -o ${name} cmd/yamon-${project}/main.go`, {
      image: imageId,
      env,
    });

    const { size } = await Deno.stat(name);
    await commitStatus?.update({
      state: "success",
      description: `${formatBytes(size)}`,
    });

    if (latest) {
      const fileData = await Deno.readFile(name);
      await upload(`yamon-${project}-${go.os}-${go.arch}`, fileData);
    }

    if (release) {
      if (client === null) {
        throw new Error(`failed to get github client`);
      }

      const fileData = await Deno.readFile(name);
      await client.uploadReleaseAsset(release, name, fileData);
    }
  }
);

export const buildAll = task("buildAll", async () => {
  await Promise.all(
    matrix.map(async ([project, variant]) => {
      await build.call({ go: variant, project });
    })
  );
});

export const github = webhook(async (event) => {
  if (event.push && event.push.head_commit) {
    for (const [project, variant] of matrix) {
      await build.spawn(
        {
          go: variant,
          project,
          latest: event.push.ref === "refs/heads/master",
        },
        {
          ref: event.push.head_commit.id,
        }
      );
    }
  } else if (event.create) {
    const client = await getClient();
    if (client === null) {
      throw new Error(`failed to get github client`);
    }

    if (event.create.ref_type === "tag" && event.create.ref.startsWith("v")) {
      const release = await client.createRelease("b1naryth1ef/yamon", {
        tag: event.create.ref,
        name: event.create.ref,
        draft: true,
      });
      for (const [project, variant] of matrix) {
        await build.spawn(
          { go: variant, project, release },
          {
            ref: event.create.ref,
          }
        );
      }
    }
  }
});
