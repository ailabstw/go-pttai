Pod::Spec.new do |spec|
  spec.name         = 'Gptt'
  spec.version      = '{{.Version}}'
  spec.license      = { :type => 'GNU Lesser General Public License, Version 3.0' }
  spec.homepage     = 'https://gitlab.corp.ailabs.tw/ptt.ai/go-pttai'
  spec.authors      = { {{range .Contributors}}
		'{{.Name}}' => '{{.Email}}',{{end}}
	}
  spec.summary      = 'iOS Pttai Client'
  spec.source       = { :git => 'https://gitlab.corp.ailabs.tw/ptt.ai/go-pttai.git', :commit => '{{.Commit}}' }

	spec.platform = :ios
  spec.ios.deployment_target  = '9.0'
	spec.ios.vendored_frameworks = 'Frameworks/Gptt.framework'

	spec.prepare_command = <<-CMD
    curl https://gpttstore.blob.core.windows.net/builds/{{.Archive}}.tar.gz | tar -xvz
    mkdir Frameworks
    mv {{.Archive}}/Gptt.framework Frameworks
    rm -rf {{.Archive}}
  CMD
end
