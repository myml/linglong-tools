outputDir=pkg/apiserver
rm -rf $outputDir || true
swaggerFile="./tools/api/client_swagger.json"
openapi-generator-cli generate --skip-validate-spec \
    -g go -o $outputDir -i $swaggerFile \
    --openapi-normalizer KEEP_ONLY_FIRST_TAG_IN_OPERATION=true \
    --additional-properties="withGoMod=false,packageName=apiserver"