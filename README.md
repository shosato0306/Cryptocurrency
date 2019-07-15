# Cryptocurrency

## Overview

This service provides a currency automatic trading function based on technical analysis.

![image](https://user-images.githubusercontent.com/38198918/58756691-c0c04680-8538-11e9-81c3-2b23d452c039.png)

## Environment

golang 1.12.4

## Functions

### View chart

You can see the candlestick chart according to the selected period.

![image](https://user-images.githubusercontent.com/38198918/61198305-49bfc580-a714-11e9-9022-6ad051daade2.png)

The candlestick period can be selected from the following four. 

- 1 second
- 1 minute
- 1 hour
- 1 day

### View indicator

You can also see the inditator on the price of the cryptocurrency.
If necessary, you can change the values of the parameters used in each indicator.

![Chart](https://user-images.githubusercontent.com/38198918/61198263-0b2a0b00-a714-11e9-817f-34a4c9b90e5c.png)

This application supports the following indicators.

- SMA ・・・ Simple Moving Average
- Ema ・・・ Exponentioal Moving Average
- BBand ・・・ Bollinger Band
- Ichimoku ・・・ Ichimoku Cloud
- Rsi ・・・ Relative Strength index
- MACD ・・・ Moving Average Convergence/Devergence Tradeing Method
- HV ・・・ Historical Volatility

### Backtest

You can use past data to simulate how much performance was obtained over a given period of time.  
The system chooses the most appropriate indicator by using backtesting.

![image](https://user-images.githubusercontent.com/38198918/58756940-a12b1d00-853c-11e9-89ef-c49bed020ded.png)

### Trading

In this service, you can actually trade using the liquid by quoine API. 
In order to trade, it is necessary to create an account.  
https://www.liquid.com/ja/
