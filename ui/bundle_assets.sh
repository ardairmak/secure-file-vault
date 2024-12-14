#!/bin/bash

echo "Generating bundled resources..."

# Set output file in current directory since we're in ui/
output_file="bundled.go"

# Remove existing bundled.go if it exists
rm -f "$output_file"

# Create fresh bundled.go file
echo "package ui" > $output_file
echo "" >> $output_file
echo "import (" >> $output_file
echo "    \"fyne.io/fyne/v2\"" >> $output_file
echo ")" >> $output_file
echo "" >> $output_file

# Bundle all assets from parent directory
for file in $(find ../assets -type f); do
    resource_name=$(basename "$file" | sed -e 's/\./_/g')
    fyne bundle -append -package ui -name "resource$resource_name" -o $output_file "$file"
done

# Add Resources map after all resources are defined
echo "" >> $output_file
echo "// Auto-generated bundled resources" >> $output_file
echo "var Resources = map[string]*fyne.StaticResource{" >> $output_file

# Add entries to Resources map
for file in $(find ../assets -type f); do
    resource_name=$(basename "$file" | sed -e 's/\./_/g')
    echo "    \"$resource_name\": resource$resource_name," >> $output_file
done

# Close the Resources map
echo "}" >> $output_file

echo "Bundling complete. Resources are available in bundled.go"