package main

import (
    "os"
    "fmt"
    "flag"
    "bufio"
    "strings"
    "sona/core"
)

type ServiceConfig struct {
    serviceKey string
    version uint
    conf map[string]string
}

func parseCommand(line string) []string {
    result := make([]string, 0)
    line = strings.Trim(line, " ")
    line_attrs := strings.Split(line, " ")
    for _, attr := range line_attrs {
        result = append(result, strings.Trim(attr, " "))
    }
    return result
}

func get(serviceKey string) *ServiceConfig {
    fmt.Println()
    serviceConf := &ServiceConfig{
        serviceKey:"richer.coolguy.leechanx",
        version:123,
        conf:map[string]string{"lover.name": "jelly claire", "money.value": "100000000", "today.age": "27"},
    }

    fmt.Printf("%-60s", "service key")
    fmt.Printf(" %s\n", serviceConf.serviceKey)
    fmt.Printf("%-60s", "version")
    fmt.Printf(" %d\n", serviceConf.version)
    fmt.Println()
    for confKey, confValue := range serviceConf.conf {
        fmt.Printf("%-60s", confKey)
        fmt.Printf(" %s\n", confValue)
    }
    fmt.Println()
    return serviceConf
}

func add(serviceKey string) {
    confKeys := make([]string, 0)
    confValues := make([]string, 0)
    for {
        var confKey string
        fmt.Printf("input a configure key (q to exit): ")
        fmt.Scanf("%s", &confKey)
        if confKey == "q" {
            break
        }
        if !core.IsValidityConfKey(confKey) {
            fmt.Println("key format error")
            continue
        }
        //check format
        fmt.Printf("input the configure value: ")
        var confValue string
        fmt.Scanf("%s", &confValue)

        confKeys = append(confKeys, confKey)
        confValues = append(confValues, confValue)
    }

    //TODO: send add request
}

func update(serviceKey string) {
    serviceConf := get(serviceKey)

    fmt.Println(">> command mode ")
    fmt.Println(">>  ")
    fmt.Println(">> add key value: add or update key value to this service")
    fmt.Println(">> del key: delete key from this service")
    fmt.Println(">> quit: finish and leave command mode")
    fmt.Println(">> ")

    reader := bufio.NewReader(os.Stdin)
    var line string
    for {
        fmt.Printf(">> ")
        line, _ = reader.ReadString('\n')
        line = strings.TrimSuffix(line, "\n")
        command := parseCommand(line)
        if len(command) == 0 {
            continue
        }

        if command[0] == "quit" {
            break
        }

        if command[0] == "add" {
            if len(command) != 3 {
                fmt.Println(">> add format error")
            } else {
                key, value := command[1], command[2]
                serviceConf.conf[key] = value
                fmt.Println(">> ok")
            }
        } else if command[0] == "del" {
            if len(command) != 2 {
                fmt.Println(">> del format error")
            } else {
                key := command[1]
                if _, ok := serviceConf.conf[key];ok {
                    delete(serviceConf.conf, key)
                    fmt.Println(">> ok")
                } else {
                    fmt.Println(">> no this key")
                }
            }
        }
    }

    fmt.Println()
    for confKey, confValue := range serviceConf.conf {
        fmt.Printf("%-60s", confKey)
        fmt.Printf(" %s\n", confValue)
    }
    fmt.Println()

    fmt.Printf("Submit? (y/n): ")
    var ensure string
    fmt.Scanf("%s", &ensure)
    if ensure == "y" || ensure == "Y" {
        //TODO
    }
}

func main() {
    host := flag.String("host", "", "admin server ip")
    port := flag.Uint("port", 0, "admin server port")
    operation := flag.String("operation", "get", "[get],[add] or [update] configures")

    flag.Parse()

    if *host == "" || *port == 0 {
        fmt.Println("no host or port is specified")
        return
    }
    if *operation != "get" && *operation != "add" && *operation != "update" {
        fmt.Println("only support operations: [get],[add] or [update]")
        return
    }

    //connect to admin server

    //ui
    var serviceKey string
    for {
        fmt.Printf("input service key: ")
        fmt.Scanf("%s", &serviceKey)
        //check format
        if !core.IsValidityServiceKey(serviceKey) {
            fmt.Println("service key format error")
        } else {
            break
        }
    }

    if *operation == "get" {
        get(serviceKey)
    } else if *operation == "add" {
        add(serviceKey)
    } else {
        update(serviceKey)
    }
}