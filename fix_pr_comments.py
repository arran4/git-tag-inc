import re

with open('config.go', 'r') as f:
    content = f.read()

# We need to change LoadConfig to find file up the tree, and change parsing to custom simple format.
# Wait, actually, let's implement the logic in Go using python just for writing the file.
