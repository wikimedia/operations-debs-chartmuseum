package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/spf13/afero"
	"helm.sh/chartmuseum/pkg/config"
)

type usageStringMap map[string]string

type errorableFlag interface {
	ApplyWithError(*flag.FlagSet) error
}

// getFlagSet is used to collect usage strings from all flags
// without havint to know their type
func getFlagSet() *flag.FlagSet {
	set := flag.NewFlagSet("", flag.ContinueOnError)
	for _, f := range config.CLIFlags {
		if ef, ok := f.(errorableFlag); ok {
			if err := ef.ApplyWithError(set); err != nil {
				log.Fatalln(err)
			}
		} else {
			f.Apply(set)
		}
	}
	return set
}

func getUsageStrings(c *config.Config) usageStringMap {
	flagSet := getFlagSet()

	var m usageStringMap
	m = make(usageStringMap)

	for _, cv := range c.AllKeys() {
		cliFlagName := config.GetCLIFlagFromVarName(cv)
		if cliFlagName == "" {
			log.Fatalln("No CLI flag associated with a config var:", cv)
		}
		f := flagSet.Lookup(cliFlagName)
		m[cv] = f.Usage
	}
	return m
}

func main() {
	c := config.NewConfig()
	usageStrings := getUsageStrings(c)
	allKeys := c.AllKeys()
	sort.Strings(allKeys)
	memFS := afero.NewMemMapFs()
	c.Viper.SetFs(memFS)
	c.WriteConfigAs("/tmp/fooo.yaml")
	b, err := afero.ReadFile(memFS, "/tmp/fooo.yaml")
	if err != nil {
		log.Fatal(err)
	}

	for _, line := range strings.Split(string(b), "\n") {
		for _, key := range allKeys {
			usage := usageStrings[key]
			if strings.HasPrefix(line, strings.Split(key, ".")[0]) {
				fmt.Printf("# %s: %s\n", key, usage)
			}
		}
		fmt.Println(line)
	}
}
