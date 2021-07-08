# Utility for Project Coconut Server/Desktop

[![CI](https://github.com/jaeha-choi/Proj_Coconut_Utility/actions/workflows/CI.yml/badge.svg)](https://github.com/jaeha-choi/Proj_Coconut_Utility/actions/workflows/CI.yml)
[![codecov](https://codecov.io/gh/jaeha-choi/Proj_Coconut_Utility/branch/master/graph/badge.svg?token=OO62TDTYH2)](https://codecov.io/gh/jaeha-choi/Proj_Coconut_Utility)

### `log` package

Contains simple wrapper methods to support level-based logging.

#### Usage

1. Import package: `import "github.com/jaeha-choi/Proj_Coconut_Utility/log"`
2. Initialize the logger: `log.Init(os.Stdout, log.DEBUG)`
3. Use one of the level to log: `log.Error(err)`

### `util` package

Contains utility methods for sending/receiving packets and defined status codes.
