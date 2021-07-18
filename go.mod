module github.com/jaeha-choi/Proj_Coconut_Desktop

go 1.16

require (
	github.com/gotk3/gotk3 v0.6.1
	github.com/jaeha-choi/Proj_Coconut_Utility v0.0.0-20210705231131-ec06b1d1b8e2
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/jaeha-choi/Proj_Coconut_Utility => ./pkg
