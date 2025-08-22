#!/bin/bash
# Complete benchmark orchestration - isolated competitor execution

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Starting Complete Benchmark Suite${NC}"
echo "======================================"

# Create timestamp for this benchmark run
TIMESTAMP=$(date +"%Y-%m-%dT%H-%M-%S")
VERSION_DIR="data/benchmark_runs/$TIMESTAMP"
mkdir -p "$VERSION_DIR"

echo -e "${GREEN}📂 Created benchmark version: $TIMESTAMP${NC}"

# Function to run competitor with memory monitoring
run_competitor() {
    local name=$1
    local dir=$2
    local script=$3
    local output_file=$4
    
    echo -e "\n${YELLOW}==================================================${NC}"
    echo -e "${YELLOW}🔬 RUNNING: $name${NC}"
    echo -e "${YELLOW}==================================================${NC}"
    
    cd "$dir"
    
    # Start the competitor process in background
    node "$script" &
    local competitor_pid=$!
    
    # Start memory monitoring
    node mem-monitor.js $competitor_pid &
    local monitor_pid=$!
    
    # Wait for competitor to finish
    wait $competitor_pid
    local exit_code=$?
    
    # Stop memory monitoring
    echo "" | node mem-monitor.js $competitor_pid > memory_stats.json 2>/dev/null || true
    kill $monitor_pid 2>/dev/null || true
    
    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}✅ $name completed successfully${NC}"
        
        # Copy results to version directory
        if [ -f "$output_file" ]; then
            cp "$output_file" "../../$VERSION_DIR/$(basename $output_file)"
            echo -e "${GREEN}📋 Results saved to $VERSION_DIR/$(basename $output_file)${NC}"
        else
            echo -e "${RED}⚠️ Warning: $output_file not found${NC}"
        fi
    else
        echo -e "${RED}❌ $name failed with exit code $exit_code${NC}"
    fi
    
    cd - > /dev/null
}

# 1. Run MArrayCRDT simulation (Go)
echo -e "\n${YELLOW}==================================================${NC}"
echo -e "${YELLOW}🔬 RUNNING: MArrayCRDT (Go)${NC}"
echo -e "${YELLOW}==================================================${NC}"

cd benchmarks
go run marraycrdt_simulation.go
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ MArrayCRDT completed successfully${NC}"
    cp marraycrdt_results.csv "../$VERSION_DIR/"
    echo -e "${GREEN}📋 Results saved to $VERSION_DIR/marraycrdt_results.csv${NC}"
else
    echo -e "${RED}❌ MArrayCRDT failed${NC}"
    exit 1
fi
cd ..

# 2. Run all Node.js competitors in isolation with cooldown between runs
run_competitor "Automerge" "competitors/automerge" "simulation.js" "automerge_results.csv"
echo -e "${BLUE}⏳ Cooling down for 60 seconds to reset Node.js state...${NC}"
sleep 60

run_competitor "Yjs" "competitors/yjs" "simulation.js" "yjs_results.csv"
echo -e "${BLUE}⏳ Cooling down for 60 seconds to reset Node.js state...${NC}"
sleep 60

run_competitor "Loro (Text)" "competitors/loro" "simulation.js" "loro_results.csv"
echo -e "${BLUE}⏳ Cooling down for 60 seconds to reset Node.js state...${NC}"
sleep 60

run_competitor "Loro (Array)" "competitors/loro" "array_simulation.js" "loro_array_results.csv"
echo -e "${BLUE}⏳ Cooling down for 60 seconds to reset Node.js state...${NC}"
sleep 60

run_competitor "Baseline" "competitors/baseline" "simulation.js" "baseline_results.csv"

# 3. Create consolidated competitor results
echo -e "\n${BLUE}📊 Consolidating competitor results...${NC}"

cd "$VERSION_DIR"
echo "system,operations,time_ms,ops_per_sec,memory_mb,final_length" > competitors_comparison.csv

# Consolidate all competitor data
for file in automerge_results.csv yjs_results.csv loro_results.csv loro_array_results.csv baseline_results.csv; do
    if [ -f "$file" ]; then
        # Get system name from filename
        system_name=$(echo "$file" | sed 's/_results.csv$//' | sed 's/loro_array/LoroArray/' | sed 's/automerge/Automerge/' | sed 's/yjs/Yjs/' | sed 's/loro/Loro/' | sed 's/baseline/Baseline/')
        
        # Skip header and copy data directly (system name already included)
        tail -n +2 "$file" | while read line; do
            if [ ! -z "$line" ] && [[ "$line" == *","* ]]; then
                echo "$line" >> competitors_comparison.csv
            fi
        done
        
        echo -e "${GREEN}  ✅ Processed: $file${NC}"
    else
        echo -e "${RED}  ⚠️ Missing: $file${NC}"
    fi
done

cd - > /dev/null

# 4. Update versions list for web UI
echo -e "\n${BLUE}📝 Updating versions list...${NC}"

VERSIONS_FILE="data/available_versions.json"

# Create versions array if it doesn't exist
if [ ! -f "$VERSIONS_FILE" ]; then
    echo "[]" > "$VERSIONS_FILE"
fi

# Count files in version directory
FILE_COUNT=$(ls -1 "$VERSION_DIR" | wc -l)

# Create new version entry
NEW_VERSION=$(cat <<EOF
{
  "version": "$TIMESTAMP",
  "path": "$VERSION_DIR",
  "created": "$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")",
  "method": "bash_isolated",
  "fileCount": $FILE_COUNT,
  "description": "Bash-orchestrated isolated benchmarks"
}
EOF
)

# Use Node.js to properly handle JSON
node -e "
const fs = require('fs');
const versionsFile = '$VERSIONS_FILE';
const newVersion = $NEW_VERSION;

let versions = [];
if (fs.existsSync(versionsFile)) {
    try {
        versions = JSON.parse(fs.readFileSync(versionsFile, 'utf8'));
    } catch (error) {
        console.error('Error parsing existing versions, creating new array');
        versions = [];
    }
}

// Add new version at the beginning
versions.unshift(newVersion);

// Keep only last 10 versions
versions = versions.slice(0, 10);

// Write back to file
fs.writeFileSync(versionsFile, JSON.stringify(versions, null, 2));
console.log('Updated versions file with ' + versions.length + ' entries');
"

echo -e "${GREEN}✅ Updated versions list${NC}"

# Results are already in the version directory, no need for additional archiving

echo -e "\n${GREEN}🎯 Complete benchmark suite finished!${NC}"
echo -e "${GREEN}📊 Results available in: $VERSION_DIR${NC}"
echo -e "${GREEN}🌐 View results at: http://localhost:3000${NC}"

echo -e "\n${BLUE}📋 Summary:${NC}"
echo "  - MArrayCRDT: $([ -f "$VERSION_DIR/marraycrdt_results.csv" ] && echo "✅" || echo "❌")"
echo "  - Automerge: $([ -f "$VERSION_DIR/automerge_results.csv" ] && echo "✅" || echo "❌")"
echo "  - Yjs: $([ -f "$VERSION_DIR/yjs_results.csv" ] && echo "✅" || echo "❌")"
echo "  - Loro: $([ -f "$VERSION_DIR/loro_results.csv" ] && echo "✅" || echo "❌")"
echo "  - LoroArray: $([ -f "$VERSION_DIR/loro_array_results.csv" ] && echo "✅" || echo "❌")"
echo "  - Baseline: $([ -f "$VERSION_DIR/baseline_results.csv" ] && echo "✅" || echo "❌")"
echo "  - Consolidated: $([ -f "$VERSION_DIR/competitors_comparison.csv" ] && echo "✅" || echo "❌")"