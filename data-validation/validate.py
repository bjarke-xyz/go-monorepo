import json
from typing import List


def validate_dynamo(ok_unleaded95_dates: List[str], ok_diesel_dates: List[str], ok_octane100_dates: List[str]):
    ddb_unleaded95_dates = set()
    ddb_diesel_dates = set()
    ddb_octane100_dates = set()
    with open('datasets/2022-04-30/dynamo_s3_export.json') as f:
        lines = f.readlines()
        for line in lines:
            ddb_item = json.loads(line)
            fueltype = ddb_item['Item']['PK']['S'].split("#")[1]
            date = ddb_item['Item']['SK']['S'].split("#")[1]
            if (fueltype == 'Unleaded95'):
                ddb_unleaded95_dates.add(date)
            elif (fueltype == 'Diesel'):
                ddb_diesel_dates.add(date)
            elif (fueltype == 'Octane100'):
                ddb_octane100_dates.add(date)

    for date in ok_unleaded95_dates:
        if date not in ddb_unleaded95_dates:
            print('UNLEADED95 ok date', date, 'not found in ddb')
    for date in ok_diesel_dates:
        if date not in ddb_diesel_dates:
            print('DIESEL ok date', date, 'not found in ddb')
    for date in ok_octane100_dates:
        if date not in ddb_octane100_dates:
            print('OCTANE100 ok date', date, 'not found in ddb')


def get_ok_dates(file) -> List[str]:
    dates = []
    with open(file, 'r') as f:
        jsonStr = f.read()
    ok_dict = json.loads(jsonStr)
    for item in ok_dict['historik']:
        dates.append(item['dato'])
    return dates


def main():
    files = ['datasets/2022-04-30/ok_unleaded95.json',
             'datasets/2022-04-30/ok_diesel.json', 'datasets/2022-04-30/ok_octane100.json']
    unleaded95_dates = get_ok_dates(files[0])
    diesel_dates = get_ok_dates(files[1])
    octane100_dates = get_ok_dates(files[2])

    validate_dynamo(
        unleaded95_dates, diesel_dates, octane100_dates)


if __name__ == "__main__":
    main()
