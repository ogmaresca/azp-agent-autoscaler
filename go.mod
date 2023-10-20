module github.com/ogmaresca/azp-agent-autoscaler

go 1.17

// Remove glog, which writes to file
// https://github.com/istio/istio/blob/master/go.mod
replace github.com/golang/glog => github.com/istio/glog v0.0.0-20190424172949-d7cfb6fa2ccd

replace k8s.io/klog => github.com/istio/klog v0.0.0-20190424230111-fb7481ea8bcf

require (
	github.com/jinzhu/copier v0.0.0-20190625015134-976e0346caa8
	github.com/prometheus/client_golang v1.0.0
	github.com/sirupsen/logrus v1.4.2
	k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apimachinery v0.15.7
	k8s.io/client-go v11.0.0+incompatible
)

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/beorn7/perks v1.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/json-iterator/go v1.1.6 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.1 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90 // indirect
	github.com/prometheus/common v0.4.1 // indirect
	github.com/prometheus/procfs v0.0.2 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	go.uber.org/atomic v1.5.0 // indirect
	go.uber.org/multierr v1.3.0 // indirect
	go.uber.org/tools v0.0.0-20190618225709-2cfd321de3ee // indirect
	go.uber.org/zap v1.14.0 // indirect
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4 // indirect
	golang.org/x/lint v0.0.0-20190930215403-16217165b5de // indirect
	golang.org/x/net v0.0.0-20190812203447-cdfb69ac37fc // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/sys v0.0.0-20190422165155-953cdadca894 // indirect
	golang.org/x/text v0.3.1-0.20181227161524-e6919f6577db // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	golang.org/x/tools v0.0.0-20191029190741-b9c20aec41a5 // indirect
	google.golang.org/appengine v1.4.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.2.4 // indirect
	honnef.co/go/tools v0.0.1-2019.2.3 // indirect
	k8s.io/klog v0.3.3 // indirect
	k8s.io/utils v0.0.0-20190607212802-c55fbcfc754a // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)
