import { isAfter, addMinutes } from "date-fns";
export interface CacheData<T> {
  data: T;
  expiryDate: Date;
}

export class Cache<T> {
  private readonly cache: Map<string, CacheData<T>>;
  constructor() {
    this.cache = new Map<string, CacheData<T>>();
  }
  insert(key: string, data: T, ttlMinutes: number = 30): void {
    const expiryDate = addMinutes(new Date(), ttlMinutes);
    this.cache.set(key, {
      data,
      expiryDate,
    });
  }

  get(key: string): T | null {
    const entry = this.cache.get(key);
    if (!entry) {
      return null;
    }

    const now = new Date();
    if (isAfter(now, entry.expiryDate)) {
      this.cache.delete(key);
      return null;
    }

    return entry.data;
  }

  delete(key: string): void {
    this.cache.delete(key);
  }
}
