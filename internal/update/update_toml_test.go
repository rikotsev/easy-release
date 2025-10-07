package update

import (
	"fmt"
	"github.com/rikotsev/easy-release/internal/config"
)

func (suite *UpdateTestSuite) TestTomlUpdate() {
	tests := []struct {
		oldContent string
		newVersion string
		newContent string
		cfg        config.Update
	}{
		{
			oldContent: `
[tool.poetry]
name = "test-python-be"
version = "0.1.0"
`,
			newVersion: "1.0.0",
			newContent: `
[tool.poetry]
name = "test-python-be"
version = "1.0.0"
`,
			cfg: config.Update{
				Kind:     config.UpdateKindToml,
				TomlPath: "3",
			},
		},
		{
			oldContent: `[tool.poetry]
name = "test-python-be"
version = "0.1.0"
description = "PVID auto repo"
authors = ["John Doe <j.doe@example.com>"]
readme = "README.md"
package-mode = false


[tool.poetry.dependencies]
python = "^3.12"
pydantic = "^2.10.3"


# dependencies used for development
[tool.poetry.group.dev.dependencies]
pre-commit = "^4.0.1"
opencv-python = "^4.10.0.84"


[tool.poetry.group.biorec.dependencies]
onnx2torch = "^1.5.15"
onnxruntime = "^1.20.1"
grad-cam = "^1.5.4"
codetiming = "^1.4.0"
requests = "^2.32.3"
omegaconf = "^2.3.0"
hydra-core = "^1.3.2"
gdown = "^5.2.0"
python-json-logger = "^3.2.1"
timm = "^1.0.12"
transformers = "^4.47.0"
matplotlib = "^3.10.0"
tensorly = "^0.9.0"
insightface = "^0.7.3"
iglovikov-helper-functions = "^0.0.53"
pymongo = "^4.10.1"
albumentations = "^1.4.22"
midv500models = "^0.0.2"
scikit-image = "^0.25.0"
glasses-detector = "^1.0.1"
tensorboard = "^2.18.0"


[tool.poetry.group.api.dependencies]
aiohttp = "^3.11.11"
aiortc = "^1.9.0"
python-swiftclient = "^4.6.0"
aiohttp-apispec = {git="https://github.com/maximdanilchenko/aiohttp-apispec", rev="3232c78"}


[tool.poetry.group.tests.dependencies]
confluent-kafka = "^2.6.1"
psutil = "^6.1.0"


[tool.poetry.group.data_model.dependencies]
aiortc = "^1.9.0"
python-dateutil = "^2.9.0.post0"


[tool.poetry.group.kafka.dependencies]
aiokafka = "^0.12.0"


# dependencies used in docker only
[tool.poetry.group.docker]
optional = true

[tool.poetry.group.docker.dependencies]
opencv-python-headless = "^4.10.0.84"


[build-system]
requires = ["poetry-core"]
build-backend = "poetry.core.masonry.api"
`,
			newVersion: "1.2.3",
			newContent: `[tool.poetry]
name = "test-python-be"
version = "1.2.3"
description = "PVID auto repo"
authors = ["John Doe <j.doe@example.com>"]
readme = "README.md"
package-mode = false


[tool.poetry.dependencies]
python = "^3.12"
pydantic = "^2.10.3"


# dependencies used for development
[tool.poetry.group.dev.dependencies]
pre-commit = "^4.0.1"
opencv-python = "^4.10.0.84"


[tool.poetry.group.biorec.dependencies]
onnx2torch = "^1.5.15"
onnxruntime = "^1.20.1"
grad-cam = "^1.5.4"
codetiming = "^1.4.0"
requests = "^2.32.3"
omegaconf = "^2.3.0"
hydra-core = "^1.3.2"
gdown = "^5.2.0"
python-json-logger = "^3.2.1"
timm = "^1.0.12"
transformers = "^4.47.0"
matplotlib = "^3.10.0"
tensorly = "^0.9.0"
insightface = "^0.7.3"
iglovikov-helper-functions = "^0.0.53"
pymongo = "^4.10.1"
albumentations = "^1.4.22"
midv500models = "^0.0.2"
scikit-image = "^0.25.0"
glasses-detector = "^1.0.1"
tensorboard = "^2.18.0"


[tool.poetry.group.api.dependencies]
aiohttp = "^3.11.11"
aiortc = "^1.9.0"
python-swiftclient = "^4.6.0"
aiohttp-apispec = {git="https://github.com/maximdanilchenko/aiohttp-apispec", rev="3232c78"}


[tool.poetry.group.tests.dependencies]
confluent-kafka = "^2.6.1"
psutil = "^6.1.0"


[tool.poetry.group.data_model.dependencies]
aiortc = "^1.9.0"
python-dateutil = "^2.9.0.post0"


[tool.poetry.group.kafka.dependencies]
aiokafka = "^0.12.0"


# dependencies used in docker only
[tool.poetry.group.docker]
optional = true

[tool.poetry.group.docker.dependencies]
opencv-python-headless = "^4.10.0.84"


[build-system]
requires = ["poetry-core"]
build-backend = "poetry.core.masonry.api"
`,
			cfg: config.Update{
				Kind:     config.UpdateKindToml,
				TomlPath: "2",
			},
		},
	}

	for idx, testCase := range tests {
		suite.Run(fmt.Sprintf("testing toml update: [%d]", idx), func() {
			updater := updateToml{
				cfg: testCase.cfg,
			}

			actual, err := updater.Run([]byte(testCase.oldContent), testCase.newVersion)
			suite.NoError(err)

			suite.Equal(testCase.newContent, string(actual))
		})
	}
}
