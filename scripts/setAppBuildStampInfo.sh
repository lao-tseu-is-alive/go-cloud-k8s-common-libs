#!/bin/bash
SOURCE_CODE=pkg/version/version.go
if [ -f "$SOURCE_CODE" ]; then
  NOW=$(date +%Y-%m-%dT%T)
  REVISION=$(git describe --dirty --always)
  sed -i "s/BuildStamp = \"unknown\"/BuildStamp = \"$NOW\"/" $SOURCE_CODE
  sed -i "s/REVISION   = \"unknown\"/REVISION = \"$REVISION\"/" $SOURCE_CODE
  echo "## Extracting app name and version from code in ${SOURCE_CODE}"
  APP_NAME=$(grep -E 'APP\s+=' $SOURCE_CODE| awk '{ print $3 }'  | tr -d '"')
  APP_VERSION=$(grep -E 'VERSION\s+=' $SOURCE_CODE| awk '{ print $3 }'  | tr -d '"')
  APP_REVISION=$(grep -E 'REVISION\s+=' $SOURCE_CODE| awk '{ print $3 }'  | tr -d '"')
  APP_BuildStamp=$(grep -E 'BuildStamp\s+=' $SOURCE_CODE| awk '{ print $3 }'  | tr -d '"')
  echo "## Found APP: ${APP_NAME}, VERSION: ${APP_VERSION}, REVISION: ${APP_REVISION}, BuildStamp: ${APP_BuildStamp}  in source file ${SOURCE_CODE}"
else
  echo "## 💥💥 ERROR: ${SOURCE_CODE} was not found !"
fi
