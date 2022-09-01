go get -u github.com/AanZee/goimportssort

# only change files different from develop branch
for entry in `git diff --name-only develop . | grep '\.go$'`; do
    echo $entry
    if grep -q "DO NOT EDIT" "$entry"; then
      continue
    fi
    goimportssort -w -local github.com/bnb-chain/ $entry
done

# change the whole file
#for entry in `find . -name "*.go"`; do
#    echo $entry
#    if grep -q "DO NOT EDIT" "$entry"; then
#      echo "xxxxxxxx=================================="
#      continue
#    fi
#    goimportssort -w -local github.com/bnb-chain/ $entry
#done
