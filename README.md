# cmc

Super-compact CLI for [Coinmarketcap API](https://coinmarketcap.com/api/documentation/v1/).

## Install

To install latest release for Linux:

```sh
wget -O cmc https://github.com/t0mk/cmc/releases/latest/download/cmc-linux-amd64 && chmod +x cmc && sudo cp cmc /usr/local/bin/
```

.. for MacOS:

```sh
wget -O cmc https://github.com/t0mk/cmc/releases/latest/download/cmc-darwin-amd64 && chmod +x cmc && sudo cp cmc /usr/local/bin/
```

## Usage

Export your Coinmarketcap API key in envvar `CMC_KEY`.

Get info about price of 1 RPL in ETH:
```sh
$ cmc v2/tools/price-conversion.amount=1,convert=eth,symbol=rpl
[
  {
    "amount": 1,
    "id": 2943,
    "last_updated": "2024-01-12T15:42:00.000Z",
    "name": "Rocket Pool",
    "quote": {
      "ETH": {
        "last_updated": "2024-01-12T15:43:00.000Z",
        "price": 0.013519053186561155
      }
    },
    "symbol": "RPL"
  }
]

```
Do the same but in more compact call:
```
$ cmc v2/t/p.s=rpl,convert=eth,a=1
```



Run `cmc` without arguments for help.





