# world-bank-etl
ETL process for publishing cleaned macroeconomic data from the World Bank API to Kaggle.com

# How to Execute the Application
### Run for 1 country and 1 indicator (GDP per capita for the USA)
```
go run ./cmd/wb/main.go -countries="USA" -indicators="NY.GDP.PCAP.CD"
```

### Run for 2 countries and 2 indicators (GDP per capita for the USA & Canada, mobile phone coverage for USA & Canada)
```
go run ./cmd/wb/main.go -countries="USA,CAN" -indicators="NY.GDP.PCAP.CD,2.0.cov.Cel"
```

### Get Help
```
go run ./cmd/wb/main.go --help 
```