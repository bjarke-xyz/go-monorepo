CREATE TABLE IF NOT EXISTS fuelprices(
    fueltype int not null,
    ts TIMESTAMP not null,
    price float NOT NULL,
    prev_prices json NOT NULL,
    PRIMARY KEY(fueltype, ts)
);
CREATE INDEX IF NOT EXISTS fuelprices_ts_index ON fuelprices(ts);
CREATE INDEX IF NOT EXISTS fuelprices_fueltype_index ON fuelprices(fueltype);