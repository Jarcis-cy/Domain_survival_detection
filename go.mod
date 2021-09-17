module Domain_survival_detection

go 1.16

require (
	github.com/schollz/progressbar/v3 v3.8.3
	github.com/widuu/gojson v0.0.0-20170212122013-7da9d2cd949b
	)

replace (
	Domain_survival_detection/pping => ../pping
	Domain_survival_detection/goWhatweb => ../goWhatweb
	Domain_survival_detection/goWhatweb/fetch => ../goWhatweb/fetch
)