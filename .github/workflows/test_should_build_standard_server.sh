if grep -q "build-standard-server: true" version; then
    echo 'result=true' >> $GITHUB_OUTPUT
else
    echo 'result=false' >> $GITHUB_OUTPUT
fi