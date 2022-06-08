/**
 * Welcome to Cloudflare Workers! This is your first worker.
 *
 * - Run `wrangler dev src/index.ts` in your terminal to start a development server
 * - Open a browser tab at http://localhost:8787/ to see your worker in action
 * - Run `wrangler publish src/index.ts --name my-worker` to publish your worker
 *
 * Learn more at https://developers.cloudflare.com/workers/
 */

import { Router } from "itty-router";
import { fetchData, getPrices, updateData } from "./handlers/prices";
import { PriceRepository } from "./lib/price-repository";
import { IttyRequest, WorkerEnv } from "./types";

const router = Router();
router.get(
  "/",
  (req: IttyRequest, env: WorkerEnv, context: EventContext<any, any, any>) =>
    getPrices(req, context, new PriceRepository(env))
);

// router.post(
//   "/prices",
//   (req: IttyRequest, env: WorkerEnv, context: EventContext<any, any, any>) =>
//     fetchData(req, context, new PriceRepository(env))
// );

// router.post(
//   "/prices/update",
//   (req: IttyRequest, env: WorkerEnv, context: EventContext<any, any, any>) =>
//     updateData(req, context, new PriceRepository(env))
// );

router.all("*", () => new Response("Not found", { status: 404 }));
export default {
  fetch: router.handle,
  async scheduled(
    event: ScheduledEvent,
    env: WorkerEnv,
    context: EventContext<any, any, any>
  ) {
    fetchData(event, context, new PriceRepository(env));
    updateData(event, context, new PriceRepository(env));
  },
};
