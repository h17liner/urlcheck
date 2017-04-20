package main

import (
	"github.com/urfave/cli"
	"net/http"
	"fmt"
	"os"
	"time"
	"bytes"
	"regexp"
	"github.com/asaskevich/govalidator"
)


func checkFlags(c *cli.Context, flags ...string) error {
	for _, flag := range flags {
		if c.String(flag) == "" {
			cli.ShowSubcommandHelp(c)
			return fmt.Errorf("%s not specified", flag)
		}
	}
	return nil
}

func getCode(body []byte) (string, error) {
	re := regexp.MustCompile(`HTTP\/[\d.]{3}\s(\d{3})`)
	status_code := string(re.FindAllSubmatch(body, 1)[0][1])

	if status_code == "" {
		return "", fmt.Errorf("Cannot find status_code")
	}

	return status_code, nil
}

func proxyRequest(c *cli.Context) error {
	if err := checkFlags(c, "url"); err != nil {
		return err
	}

	request_url := c.String("url")
	valid := govalidator.IsRequestURL(request_url)

	if valid == false {
		fmt.Printf("%v is a invalid url", request_url)
		return fmt.Errorf("%v is a invalid url", request_url)
	}
	timeout := c.Int("timeout")

	if timeout == 0 {
		timeout = 5
	}

	user_agent := c.String("user_agent")
	if user_agent == "" {
		user_agent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.12; rv:51.0) Gecko/20100101 Firefox/51.0"
	}

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}


	req_string := []byte(fmt.Sprintf("url=%s&method=GET&auth=none", request_url))
	req, err := http.NewRequest("POST", "http://hurl.eu/", bytes.NewBuffer(req_string))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", user_agent)

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(make([]byte, 0, resp.ContentLength))
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	body := buf.Bytes()
	code, err := getCode(body)
	if err != nil {
		return err
	}

	fmt.Printf(code)

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "urlcheck"
	app.Usage = "URL Check CLI"
	app.Version = "2017.4.0"
	app.Action = proxyRequest
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "url",
		},
		cli.StringFlag{
			Name: "user_agent",
		},
		cli.IntFlag{
			Name: "timeout",
		},
	}

	app.Run(os.Args)
}