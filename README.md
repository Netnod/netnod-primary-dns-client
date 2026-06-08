# netnod-primary-dns-client

Go client for [Netnod Primary DNS API](https://www.netnod.se/).

## Installation

```bash
go get github.com/netnod/netnod-primary-dns-client
```

## Use

```go
package main

import (
    "fmt"
    "log"

    netnod "github.com/netnod/netnod-primary-dns-client"
)

func main() {
    client := netnod.NewClient("", "your-api-token") // empty URL = default

    // List all zones
    zones, err := client.ListZones()

    // List zones filtered by end customer
    zones, err = client.ListZones(netnod.WithEndCustomerName("customer123"))
    if err != nil {
        log.Fatal(err)
    }
    for _, z := range zones {
        fmt.Printf("%s (serial: %d)\n", z.Name, z.NotifiedSerial)
    }

    // Fetch a zone with all records
    zone, err := client.GetZone("example.com.")
    if err != nil {
        log.Fatal(err)
    }
    for _, rrset := range zone.RRsets {
        fmt.Printf("%s %s\n", rrset.Name, rrset.Type)
    }

    // Create / update RR
    ttl := int64(3600)
    err = client.PatchZoneRRsets("example.com.", []netnod.RRset{
        {
            Name:       "www.example.com.",
            Type:       "A",
            TTL:        &ttl,
            ChangeType: "REPLACE",
            Records: []netnod.Record{
                {Content: "192.0.2.1", Disabled: false},
            },
        },
    })

    // Create zone from BIND zone file
    created, err := client.CreateZoneFromBIND(&netnod.ZoneCreateBIND{
        Name:       "example.com.",
        Zone:       "example.com.\t3600\tIN\tSOA\tns1.example.com. hostmaster.example.com. 2025110401 10800 3600 604800 3600\nexample.com.\t3600\tIN\tNS\tns1.example.com.",
        AlsoNotify: []string{"1.2.3.4"},
    })

    // Export zone in BIND format
    bindData, err := client.ExportZone("example.com.")

    // Trigger immediate DNS NOTIFY
    _, err = client.NotifyZone("example.com.")

    // DynDNS management
    labels, err := client.ListDynDNS("example.com.")
    dyndns, err := client.CreateDynDNS("example.com.", "home")
    fmt.Println(dyndns.Token) // save this token
    err = client.DeleteDynDNS("example.com.", "home")

    // ACME DNS-01 challenge management
    acmeLabels, err := client.ListACME("example.com.")
    acme, err := client.CreateACME("example.com.", "www")
    fmt.Println(acme.Token) // save this token
    err = client.DeleteACME("example.com.", "www")
}
```

## API

### Client

```go
// Create a client (empty baseURL = https://primarydnsapi.netnod.se)
client := netnod.NewClient(baseURL, token)
```

### Zones

| Method                     | Description                                                           |
| -------------------------- | --------------------------------------------------------------------- |
| `ListZones(options...)`    | List all zones; optionally pass `WithEndCustomerName(name)` to filter |
| `GetZone(zoneID)`          | Get zone with all RRsets                                              |
| `CreateZone(zone)`         | Create new zone with RRsets                                           |
| `CreateZoneFromBIND(zone)` | Create new zone from BIND zone file data                              |
| `UpdateZone(zoneID, zone)` | Update zone configuration                                             |
| `DeleteZone(zoneID)`       | Delete zone                                                           |
| `ExportZone(zoneID)`       | Export zone in BIND zone file format                                  |
| `NotifyZone(zoneID)`       | Trigger immediate DNS NOTIFY                                          |

### Records

| Method                            | Description                  |
| --------------------------------- | ---------------------------- |
| `PatchZoneRRsets(zoneID, rrsets)` | Create/update/delete records |
| `GetRRset(zoneID, name, type)`    | Get specific RRset           |

### DynDNS

| Method                        | Description                               |
| ----------------------------- | ----------------------------------------- |
| `ListDynDNS(zoneID)`          | List DynDNS-enabled labels                |
| `CreateDynDNS(zoneID, label)` | Enable DynDNS for a label (returns token) |
| `DeleteDynDNS(zoneID, label)` | Disable DynDNS for a label                |

### ACME DNS-01

| Method                      | Description                             |
| --------------------------- | --------------------------------------- |
| `ListACME(zoneID)`          | List ACME-enabled labels                |
| `CreateACME(zoneID, label)` | Enable ACME for a label (returns token) |
| `DeleteACME(zoneID, label)` | Disable ACME for a label                |

### ChangeType for PatchZoneRRsets

| Value     | Description                        |
| --------- | ---------------------------------- |
| `REPLACE` | Replace all records in RRset       |
| `DELETE`  | Delete RRset                       |
| `EXTEND`  | Add records if not already present |
| `PRUNE`   | Remove specific records from RRset |

## License

BSD 3-Clause License, see [LICENSE](LICENSE).
