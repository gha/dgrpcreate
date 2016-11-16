package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

var inputFile string
var igrp string
var apiBase string

var username string
var password string

func init() {
	flag.StringVar(&inputFile, "file", "", "Input file")
	flag.StringVar(&igrp, "igrp", "", "IGRP ID")
	flag.StringVar(&apiBase, "base", "https://api.infinitycloud.com/config/", "API base URL")
}

func main() {
	fmt.Println("Starting...")
	flag.Parse()

	err := checkInput()
	if err != nil {
		fmt.Println(err)
		return
	}

	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	username, password, err = getCredentials()
	if err != nil {
		fmt.Println(err)
		return
	}

	reader := csv.NewReader(bufio.NewReader(file))
	count := 0
	headings := []string{}
	for {
		row, err := reader.Read()
		if err == io.EOF {
			fmt.Println("End of file")
			break
		}

		count++
		if count == 1 {
			headings = row
			continue
		}

		dgrp, err := mapRow(headings, row)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error mapping row: %s. Skipping %s", err, row))
			continue
		}

		err = processDgrp(dgrp)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func checkInput() error {
	if len(inputFile) == 0 {
		return errors.New("Missing input file")
	}

	if len(igrp) == 0 {
		return errors.New("Missing IGRP ID")
	}

	_, err := os.Stat(inputFile)
	return err
}

func getCredentials() (string, string, error) {
	r := bufio.NewReader(os.Stdin)

	fmt.Println("Enter username:")
	user, err := r.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	fmt.Println("Enter password:")
	pass, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", err
	}

	return strings.TrimSpace(user), strings.TrimSpace(string(pass)), nil
}

func mapRow(headings, row []string) (map[string]string, error) {
	if len(headings) != len(row) {
		return nil, errors.New("Incorrect number of fields")
	}

	ret := make(map[string]string)
	for k, v := range row {
		ret[headings[k]] = v
	}

	return ret, nil
}

func processDgrp(dgrp map[string]string) error {
	fmt.Print(dgrp["dgrpName"] + ": ")

	apiUrl := strings.TrimSuffix(apiBase, "/") + "/v2/igrps/" + igrp + "/dgrps/"
	data := url.Values{}
	for k, v := range dgrp {
		data.Set(k, v)
	}

	apiUrl = apiUrl + "?" + data.Encode()

	req, err := http.NewRequest("POST", apiUrl, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(username, password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	fmt.Print(resp.Status)
	fmt.Print("...")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(strings.TrimSpace(string(body)))
	return nil
}
