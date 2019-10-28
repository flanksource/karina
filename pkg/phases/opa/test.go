package opa



//
// func Test(p *platform.Platform, test *console.TestResults) {

func Test() {
	kubectl := p.GetKubectl()
	var folders []string
	for _, folder := range folders {
		pass := fmt.Sprintf("%s/pass.yml", folder )

		if err := kubectl("apply -f %s", pass); err != nil {
			test.Passf("%s applied correctly", pass)
		} else {
			test.Failf("%s should have applied, but did not %v", folder, err)
		}

		fail := fmt.Sprintf("%s/fail.yml", folder )

		if err := kubectl("apply -f %s", fail); err != nil {
			test.Failf("%s should not have applied, but did", fail)
				} else {
			test.Passf("%s was not applied correctly", fail)
		}
	}

}