// Copyright 2023 The frp Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validation

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/samber/lo"

	v1 "github.com/fatedier/frp/pkg/config/v1"
)

func ValidateClientCommonConfig(c *v1.ClientCommonConfig) (Warning, error) {
	var (
		warnings Warning
		errs     error
	)
	if c.Transport.HeartbeatTimeout > 0 && c.Transport.HeartbeatInterval > 0 {
		if c.Transport.HeartbeatTimeout < c.Transport.HeartbeatInterval {
			errs = AppendError(errs, fmt.Errorf("invalid transport.heartbeatTimeout, heartbeat timeout should not less than heartbeat interval"))
		}
	}

	if !lo.FromPtr(c.Transport.TLS.Enable) {
		checkTLSConfig := func(name string, value string) Warning {
			if value != "" {
				return fmt.Errorf("%s is invalid when transport.tls.enable is false", name)
			}
			return nil
		}

		warnings = AppendError(warnings, checkTLSConfig("transport.tls.certFile", c.Transport.TLS.CertFile))
		warnings = AppendError(warnings, checkTLSConfig("transport.tls.keyFile", c.Transport.TLS.KeyFile))
		warnings = AppendError(warnings, checkTLSConfig("transport.tls.trustedCaFile", c.Transport.TLS.TrustedCaFile))
	}

	if !lo.Contains([]string{"tcp", "kcp", "quic", "websocket", "wss"}, c.Transport.Protocol) {
		errs = AppendError(errs, fmt.Errorf("invalid transport.protocol, only support tcp, kcp, quic, websocket, wss"))
	}

	for _, f := range c.IncludeConfigFiles {
		absDir, err := filepath.Abs(filepath.Dir(f))
		if err != nil {
			errs = AppendError(errs, fmt.Errorf("include: parse directory of %s failed: %v", f, err))
			continue
		}
		if _, err := os.Stat(absDir); os.IsNotExist(err) {
			errs = AppendError(errs, fmt.Errorf("include: directory of %s not exist", f))
		}
	}
	return warnings, errs
}

func ValidateAllClientConfig(c *v1.ClientCommonConfig, pxyCfgs []v1.ProxyConfigurer, visitorCfgs []v1.VisitorConfigurer) (Warning, error) {
	var warnings Warning
	if c != nil {
		warning, err := ValidateClientCommonConfig(c)
		warnings = AppendError(warnings, warning)
		if err != nil {
			return err, warnings
		}
	}

	for _, c := range pxyCfgs {
		if err := ValidateProxyConfigurerForClient(c); err != nil {
			return warnings, fmt.Errorf("proxy %s: %v", c.GetBaseConfig().Name, err)
		}
	}

	for _, c := range visitorCfgs {
		if err := ValidateVisitorConfigurer(c); err != nil {
			return warnings, fmt.Errorf("visitor %s: %v", c.GetBaseConfig().Name, err)
		}
	}
	return warnings, nil
}
