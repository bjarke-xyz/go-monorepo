export type FuelType = "Unleaded95" | "Octane100" | "Diesel";
export function fuelTypeToOkItemNumber(fuelType: FuelType): number {
  switch (fuelType) {
    case "Unleaded95":
      return 536;
    case "Octane100":
      return 533;
    case "Diesel":
      return 231;
    default:
      return 536;
  }
}

export interface OkPrices {
  historik: {
    dato: string;
    pris: number;
    prisExclAfgifterExclMoms?: number;
    prisExclAfgifterInclMoms?: number;
    prisExclMoms?: number;
    varenr?: number;
  }[];
}

export interface Price {
  date: string;
  price: number;
}

export interface DayPrices {
  today: Price;
  yesterday: Price | null;
  tomorrow: Price | null;
}
