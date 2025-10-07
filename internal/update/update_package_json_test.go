package update

import (
	"fmt"
	"github.com/rikotsev/easy-release/internal/config"
)

func (suite *UpdateTestSuite) TestPackageJsonUpdate() {
	tests := []struct {
		oldContent string
		newVersion string
		newContent string
		cfg        config.Update
	}{
		{
			oldContent: `{
	"name": "my-cool-npm-package",
	"version": "0.0.2"
}`,
			newVersion: "1.2.3",
			newContent: `{
	"name": "my-cool-npm-package",
	"version": "1.2.3"
}`,
		},
		{
			oldContent: `{
	"name": "my-cool-npm-package-without-version"
}`,
			newVersion: "1.2.3",
			newContent: `{
	"name": "my-cool-npm-package-without-version"
}`,
		},
		{
			oldContent: `{
  "name": "id-lre-front",
  "version": "0.1.85",
  "private": true,
  "dependencies": {
    "@org/app-styles": "^0.0.45",
    "@org/app-sapphire": "^0.0.70",
    "@reduxjs/toolkit": "^2.0.1",
    "@testing-library/jest-dom": "^5.17.0",
    "@testing-library/react": "^13.4.0",
    "@testing-library/user-event": "^13.5.0",
    "@types/jest": "^27.5.2",
    "@types/node": "^16.18.61",
    "@types/react": "=18.2.37",
    "@types/react-dom": "=18.2.15",
    "i18next": "^23.15.1",
    "papaparse": "^5.4.1",
    "react": "=18.2.0",
    "react-dom": "=18.2.0",
    "react-i18next": "^15.0.2",
    "react-redux": "^9.0.4",
    "react-router-dom": "^6.19.0",
    "react-scripts": "5.0.1",
    "react-transition-group": "^4.4.5",
    "sonarqube-scanner": "^4.0.1",
    "typescript": "^4.9.5",
    "web-vitals": "^2.1.4"
  },
  "scripts": {
    "sonar": "node ./scripts/sonar/cli.js",
    "sonar:scan": "node ./scripts/sonar/pipeline.js",
    "authenticate": "npx vsts-npm-auth -config .npmrc",
    "start": "react-app-rewired start",
    "build": "react-app-rewired build",
    "test": "CI=true react-app-rewired test",
    "test:watch": "react-app-rewired test",
    "eject": "react-scripts eject",
    "format": "prettier . --write",
    "lint": "eslint src --ext .js,.jsx,.ts,.tsx",
    "lint:fix": "npm run lint -- --fix",
    "start:local": "npm run build && node ./scripts/generate_config.js && npx live-server ./build --entry-file=index.html --port=3001 --no-browser"
  },
  "eslintConfig": {
    "extends": [
      "react-app",
      "react-app/jest"
    ]
  },
  "browserslist": {
    "production": [
      ">0.2%",
      "not dead",
      "not op_mini all"
    ],
    "development": [
      "last 1 chrome version",
      "last 1 firefox version",
      "last 1 safari version"
    ]
  },
  "devDependencies": {
    "@babel/preset-typescript": "^7.24.1",
    "@types/papaparse": "^5.3.14",
    "@types/react-transition-group": "^4.4.10",
    "cross-env": "^7.0.3",
    "eslint-config-prettier": "^9.1.0",
    "eslint-plugin-prettier": "^5.0.1",
    "jest-transform-stub": "^2.0.0",
    "prettier": "3.1.0",
    "prettier-plugin-multiline-arrays": "^3.0.1",
    "react-app-rewired": "^2.2.1",
    "sass": "^1.71.1",
    "sass-loader": "^14.1.1",
    "ts-jest": "^29.2.4"
  }
}`,
			newVersion: "4.5.6",
			newContent: `{
  "name": "id-lre-front",
  "version": "4.5.6",
  "private": true,
  "dependencies": {
    "@org/app-styles": "^0.0.45",
    "@org/app-sapphire": "^0.0.70",
    "@reduxjs/toolkit": "^2.0.1",
    "@testing-library/jest-dom": "^5.17.0",
    "@testing-library/react": "^13.4.0",
    "@testing-library/user-event": "^13.5.0",
    "@types/jest": "^27.5.2",
    "@types/node": "^16.18.61",
    "@types/react": "=18.2.37",
    "@types/react-dom": "=18.2.15",
    "i18next": "^23.15.1",
    "papaparse": "^5.4.1",
    "react": "=18.2.0",
    "react-dom": "=18.2.0",
    "react-i18next": "^15.0.2",
    "react-redux": "^9.0.4",
    "react-router-dom": "^6.19.0",
    "react-scripts": "5.0.1",
    "react-transition-group": "^4.4.5",
    "sonarqube-scanner": "^4.0.1",
    "typescript": "^4.9.5",
    "web-vitals": "^2.1.4"
  },
  "scripts": {
    "sonar": "node ./scripts/sonar/cli.js",
    "sonar:scan": "node ./scripts/sonar/pipeline.js",
    "authenticate": "npx vsts-npm-auth -config .npmrc",
    "start": "react-app-rewired start",
    "build": "react-app-rewired build",
    "test": "CI=true react-app-rewired test",
    "test:watch": "react-app-rewired test",
    "eject": "react-scripts eject",
    "format": "prettier . --write",
    "lint": "eslint src --ext .js,.jsx,.ts,.tsx",
    "lint:fix": "npm run lint -- --fix",
    "start:local": "npm run build && node ./scripts/generate_config.js && npx live-server ./build --entry-file=index.html --port=3001 --no-browser"
  },
  "eslintConfig": {
    "extends": [
      "react-app",
      "react-app/jest"
    ]
  },
  "browserslist": {
    "production": [
      ">0.2%",
      "not dead",
      "not op_mini all"
    ],
    "development": [
      "last 1 chrome version",
      "last 1 firefox version",
      "last 1 safari version"
    ]
  },
  "devDependencies": {
    "@babel/preset-typescript": "^7.24.1",
    "@types/papaparse": "^5.3.14",
    "@types/react-transition-group": "^4.4.10",
    "cross-env": "^7.0.3",
    "eslint-config-prettier": "^9.1.0",
    "eslint-plugin-prettier": "^5.0.1",
    "jest-transform-stub": "^2.0.0",
    "prettier": "3.1.0",
    "prettier-plugin-multiline-arrays": "^3.0.1",
    "react-app-rewired": "^2.2.1",
    "sass": "^1.71.1",
    "sass-loader": "^14.1.1",
    "ts-jest": "^29.2.4"
  }
}`,
		},
	}

	for idx, testCase := range tests {
		suite.Run(fmt.Sprintf("testing package.json update: [%d]", idx), func() {
			update := updatePackageJson{
				cfg: testCase.cfg,
			}

			actual, err := update.Run([]byte(testCase.oldContent), testCase.newVersion)
			suite.Require().NoError(err)
			suite.Require().Equal(testCase.newContent, string(actual))
		})
	}
}
