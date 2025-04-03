# Telegram Gift NFT Scanner

A utility for scanning Telegram gift NFTs and collecting statistics on models and their rarity.

## Features

- Scanning gift NFTs by specified model
- Multi-threaded data collection
- Progress display during scanning
- Statistics output in a table format, sorted by rarity
- For each model shows:
  - Number of found NFTs
  - Percentage of total
  - Model rarity

## Installation

```bash
git clone https://github.com/yourusername/gifts-scanner.git
cd gifts-scanner
go build
```

## Usage

```bash
./gifts-scanner <ModelName> <count> [threads]
```

Where:
- `ModelName` - NFT model name (e.g., HomemadeCake)
- `count` - number of NFTs to scan
- `threads` - number of threads for parallel scanning (default: 10)

Example:
```bash
./gifts-scanner HomemadeCake 1000 20
```

## Example Output

```
     Telegram Gift Scanner - Final Results     
                               
 INFO  Checked: 100 gifts, Available: 100

Model         | Count | Percent | Rarity
Frosty        | 2     | 2.0%    | 0.2%
Red Pearl     | 2     | 2.0%    | 0.5%
Pink Whirl    | 3     | 3.0%    | 1.0%
Festive Mint  | 2     | 2.0%    | 1.0%
Pistachio Sky | 4     | 4.0%    | 1.5%
Pumpkin Pie   | 4     | 4.0%    | 1.5%
...
``` 