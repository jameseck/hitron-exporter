# hitron-exporter

Metrics exporter for Hitron CGNV4-FX2 routers, which are (used to be?) distributed as part of Unitymedia/Vodafone Business cable internet.


## Running

```bash
docker run -it --rm -p 9101:80 cfstras/hitron-exporter --host --pass XYZ
```
## License

See [LICENSE.md](LICENSE.md)  
Heavily inspired by and structure stolen from [fluepke's vodafone-station-exporter](https://github.com/Fluepke/vodafone-station-exporter).  
