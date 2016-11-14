package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"

	"github.com/vpn-kill-switch/killswitch"
)

func exit1(err error) {
	fmt.Println(err)
	os.Exit(1)
}

var version string

func main() {

	var (
		ip = flag.String("ip", "", "VPN peer `IPv4`")
		e  = flag.Bool("e", false, "`Enable` load the pf rules")
		i  = flag.Bool("i", false, "`Info` print active interfaces.")
		v  = flag.Bool("v", false, fmt.Sprintf("Print version: %s", version))
	)

	flag.Parse()

	if *v {
		fmt.Printf("%s\n", version)
		os.Exit(0)
	}

	ks, err := killswitch.New(*ip)
	if err != nil {
		exit1(err)
	}

	err = ks.GetActive()
	if err != nil {
		exit1(err)
	}

	if *i {
		for k, v := range ks.UpInterfaces {
			fmt.Println(k, v)
		}
	} else if *ip == "" {
		exit1(fmt.Errorf("Please enter the VPN peer IP, use (\"%s -h\") for help.\n", os.Args[0]))
	} else if ipv4 := net.ParseIP(*ip); ipv4.To4() == nil {
		exit1(fmt.Errorf("%s is not a valid IPv4 address, use (\"%s -h\") for help.\n", *ip, os.Args[0]))
	}

	if len(ks.P2PInterfaces) == 0 {
		exit1(fmt.Errorf("No VPN interface found, verify VPN is connected, use (\"%s -h\") for help.\n", os.Args[0]))
	}

	ks.CreatePF()

	fmt.Println(ks.PFRules.String())

	usr, err := user.Current()
	if err != nil {
		exit1(err)
	}
	if err = ioutil.WriteFile(path.Join(usr.HomeDir, ".killswitch.pf.conf"),
		ks.PFRules.Bytes(),
		0644,
	); err != nil {
		exit1(err)
	}

	if *e {
		fmt.Printf("# %s\n", strings.Repeat("-", 62))
		fmt.Println("# Loading rules")
		fmt.Printf("# %s\n", strings.Repeat("-", 62))
		out, _ := exec.Command("pfctl", "-e").CombinedOutput()
		fmt.Printf("%s\n", out)
		out, _ = exec.Command("pfctl",
			"-Fa",
			"-f",
			path.Join(usr.HomeDir, ".killswitch.pf.conf")).CombinedOutput()
		fmt.Printf("%s\n", out)
		out, _ = exec.Command("pfctl", "-sr").CombinedOutput()
		fmt.Printf("%s\n", out)
	}
}