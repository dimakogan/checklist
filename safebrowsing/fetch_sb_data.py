# Source: https://github.com/Screetsec/Sudomy/issues/28
#
# This fetches the data displayed on the Google Safe Browsing Transparency Report and outputs it as a CSV
# that can be imported into a Kaggle dataset.
# The original visualization can be found here: https://transparencyreport.google.com/safe-browsing/overview

import numpy as np # linear algebra
import pandas as pd # data processing, CSV file I/O (e.g. pd.read_csv)

from datetime import datetime, date
RUN_TIME = int(datetime.utcnow().timestamp() * 1000)
START_TIME = datetime.fromtimestamp(1148194800000 // 1000)

# Here are the URL requests I found on the page: 
UNSAFE_URL = f"https://transparencyreport.google.com/transparencyreport/api/v3/safebrowsing/sites?dataset=0&series=malwareDetected,phishingDetected&start=1148194800000&end={RUN_TIME}"
NUMBER_URL = f"https://transparencyreport.google.com/transparencyreport/api/v3/safebrowsing/sites?dataset=1&series=malware,phishing&start=1148194800000&end={RUN_TIME}"
SITES_URL = f"https://transparencyreport.google.com/transparencyreport/api/v3/safebrowsing/sites?start=1148194800000&series=attack,compromised&end={RUN_TIME}"
BROWSER_WARNINGS_URL = f"https://transparencyreport.google.com/transparencyreport/api/v3/safebrowsing/warnings?dataset=users&start=1148194800000&end={RUN_TIME}&series=users"
SEARCH_WARNINGS_URL = f"https://transparencyreport.google.com/transparencyreport/api/v3/safebrowsing/warnings?dataset=search&start=1148194800000&end={RUN_TIME}&series=search"
RESPONSE_TIME_URL = f"https://transparencyreport.google.com/transparencyreport/api/v3/safebrowsing/notify?dataset=1&start=1148194800000&end={RUN_TIME}&series=response"
REINFECTION_URL = f"https://transparencyreport.google.com/transparencyreport/api/v3/safebrowsing/notify?dataset=0&start=1148194800000&end={RUN_TIME}&series=reinfect"

COLUMN_NAMES = [
    "WeekOf",
    "Malware sites detected",
    "Phishing sites detected",
    "Malware sites number",
    "Phishing sites number",
    "Attack sites",
    "Compromised sites",
    "Browser warnings",
    "Search warnings",
    "Webmaster response time",
    "Reinfection rate"
    ]

def load_dataframe():
    dates = pd.date_range(start=START_TIME, end=datetime.fromtimestamp(RUN_TIME // 1000), freq='W', normalize=True)
    df = pd.DataFrame(columns=COLUMN_NAMES)
    df["WeekOf"] = dates
    df = df.set_index("WeekOf")
    return df
    
df = load_dataframe()

import requests
import json
def fetch_as_json(url):
    r = requests.get(url)
    c = r.content
    c = c[5:]
    j = json.loads(c)
    return j[0][1]

def malware_phishing_detected(df):
    pts = fetch_as_json(UNSAFE_URL)
    for pt in pts:
        date = pd.to_datetime(pt[0], unit='ms').normalize()
        malware = pt[1][0]
        phishing = pt[1][1]
        malware = malware[0] if malware else np.NaN
        phishing = phishing[0] if phishing else np.NaN
        df[COLUMN_NAMES[1]][date] = malware
        df[COLUMN_NAMES[2]][date] = phishing
    return df

def malware_phishing_number(df):
    pts = fetch_as_json(NUMBER_URL)
    for pt in pts:
        date = pd.to_datetime(pt[0], unit='ms').normalize()
        malware = pt[1][0]
        phishing = pt[1][1]
        malware = malware[0] if malware else np.NaN
        phishing = phishing[0] if phishing else np.NaN
        df[COLUMN_NAMES[3]][date] = malware
        df[COLUMN_NAMES[4]][date] = phishing
    return df
        
def site_count(df):
    pts = fetch_as_json(SITES_URL)
    for pt in pts:
        date = pd.to_datetime(pt[0], unit='ms').normalize()
        attack = pt[1][0]
        comped = pt[1][1]
        attack = attack[0] if attack else np.NaN
        comped = comped[0] if comped else np.NaN
        df[COLUMN_NAMES[5]][date] = attack
        df[COLUMN_NAMES[6]][date] = comped
    return df
    
def browser_warnings(df):
    pts = fetch_as_json(BROWSER_WARNINGS_URL)
    for pt in pts:
        date = pd.to_datetime(pt[0], unit='ms').normalize()
        value = pt[1][0]
        value = value[0] if value else np.NaN
        df[COLUMN_NAMES[7]][date] = value
    return df
    
def search_warnings(df):
    pts = fetch_as_json(SEARCH_WARNINGS_URL)
    for pt in pts:
        date = pd.to_datetime(pt[0], unit='ms').normalize()
        value = pt[1][0]
        value = value[0] if value else np.NaN
        df[COLUMN_NAMES[8]][date] = value
    return df
    
def response_time(df):
    pts = fetch_as_json(RESPONSE_TIME_URL)
    for pt in pts:
        date = pd.to_datetime(pt[0], unit='ms').normalize()
        value = pt[1][0]
        value = value[0] if value else np.NaN
        df[COLUMN_NAMES[9]][date] = value
    return df
    
def reinfection_rate(df):
    pts = fetch_as_json(REINFECTION_URL)
    for pt in pts:
        date = pd.to_datetime(pt[0], unit='ms').normalize()
        value = pt[1][0]
        # Multiply by 100 and cast to int to save space on import.
        value = int(value[1] * 100) if value else np.NaN
        df[COLUMN_NAMES[10]][date] = value
    return df

df = malware_phishing_detected(df)
df = malware_phishing_number(df)
df = site_count(df)
df = browser_warnings(df)
df = search_warnings(df)
df = response_time(df)
df = reinfection_rate(df)
df.to_csv("data.csv", header=True, index=True, index_label="WeekOf")