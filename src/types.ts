import { Request as IttyRouterRequest } from "itty-router";

export type WorkerEnv = {
  R2_FUELPRICES: R2Bucket;
  KV_FUELPRICES: KVNamespace;
};

export type IttyRequest = Request & IttyRouterRequest;
