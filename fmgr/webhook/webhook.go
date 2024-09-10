package webhook

import (
	"frpgo/config"
	"frpgo/pkg2/utils2"

	"github.com/go-resty/resty/v2"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	whClient *resty.Client
	whUrl    string
)

func init() {
	whClient = resty.New()
}

func Setup(c config.Config) {
	whUrl = c.Webhook.Url

	logx.Debugf("Setup url: %v", whUrl)
}

func PushProxyDetail(data interface{}) {
	go webhook(whUrl, data)
}

func webhook(url string, data interface{}) error {
	logx.Debugf("PushProxyDetail url: %v, data: %v", url, data)

	r, err := post(url, data)
	if err != nil {
		rstr := utils2.PrettyJson(r)
		logx.Errorf("err: %v, resp: %v", err, rstr)
		return err
	}

	return nil
}

func post(url string, data interface{}) (r *resty.Response, e error) {
	request, err := whClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(data).
		Post(url)
	if err != nil {
		return request, err
	}

	logx.Debugf("post response : %v", string(request.Body()))

	statusCode := request.StatusCode()
	if statusCode == 0 {
		return request, err
	}

	return request, err
}
