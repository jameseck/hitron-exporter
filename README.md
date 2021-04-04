# hitron-exporter

Metrics exporter for Hitron CGNV4-FX2 routers, which are (used to be?) distributed as part of Unitymedia/Vodafone Business cable internet.


## Running

```bash
docker run -it --rm -p 9101:80 ghcr.io/cfstras/hitron-exporter:latest --host --pass XYZ
```

### docker-compose

```yaml
hitron_exporter:
  image: ghcr.io/cfstras/hitron-exporter:latest
  command:
    - --pass=mySecretPassword
#   - --user=admin
#   - --host=http://192.168.0.1:80/
  ports:
    - 9101:80
  restart: unless-stopped
```

### Example output

```bash
➜ curl localhost:9101/metrics
# HELP hitron_address Hardware and IP Addresses in labels
# TYPE hitron_address gauge
hitron_address{lan_ip="192.168.0.1/24",rf_mac="68:8F:12:34:12:34",wan_ip="84.12.34.56/21"} 1
# HELP hitron_cm_bpi_status DOCSIS Provisioning BPI Status
# TYPE hitron_cm_bpi_status gauge
hitron_cm_bpi_status{auth="authorized",tek="operational"} 1
# HELP hitron_cm_dhcp_success DOCSIS Provisioning DHCP Status
# TYPE hitron_cm_dhcp_success gauge
hitron_cm_dhcp_success 1
# HELP hitron_cm_download_config_success DOCSIS Provisioning Download CM Config File Status
# TYPE hitron_cm_download_config_success gauge
hitron_cm_download_config_success 1
# HELP hitron_cm_find_downstream_success DOCSIS Provisioning Lock Downstream Status
# TYPE hitron_cm_find_downstream_success gauge
hitron_cm_find_downstream_success 1
# HELP hitron_cm_hwinit_success DOCSIS Provisioning HWInit Status
# TYPE hitron_cm_hwinit_success gauge
hitron_cm_hwinit_success 1
# HELP hitron_cm_network_access_status DOCSIS Network Access Permission
# TYPE hitron_cm_network_access_status gauge
hitron_cm_network_access_status 1
# HELP hitron_cm_ranging_success DOCSIS Provisioning Ranging Status
# TYPE hitron_cm_ranging_success gauge
hitron_cm_ranging_success 1
# HELP hitron_cm_registration_success DOCSIS Provisioning Registration Status
# TYPE hitron_cm_registration_success gauge
hitron_cm_registration_success 1
# HELP hitron_info_uptime System uptime
# TYPE hitron_info_uptime counter
hitron_info_uptime 516435
# HELP hitron_login_success_bool 1 if the login was successful
# TYPE hitron_login_success_bool gauge
hitron_login_success_bool 1
# HELP hitron_scrape_time Time the scrape run took
# TYPE hitron_scrape_time gauge
hitron_scrape_time 4.046142458
# HELP hitron_traffic Basic traffic counters. if=wan/lan, dir=send/recv.
# TYPE hitron_traffic counter
hitron_traffic{dir="recv",if="lan"} 1.015283712e+09
hitron_traffic{dir="recv",if="wan"} 9.4384422912e+08
hitron_traffic{dir="send",if="lan"} 1.75019917312e+09
hitron_traffic{dir="send",if="wan"} 6.3736643584e+08
# HELP hitron_version Versions in labels
# TYPE hitron_version gauge
hitron_version{hw_version="1A",serial="VCAP12345678",sw_version="4.12.34.567-XX-YYY"} 1
```
## License

See [LICENSE.md](LICENSE.md)  
Heavily inspired by and structure stolen from [fluepke's vodafone-station-exporter](https://github.com/Fluepke/vodafone-station-exporter).  
