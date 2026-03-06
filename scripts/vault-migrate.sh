#!/bin/bash
# Vault Migration Script v2
# Normalizes vault notes to follow the updated conventions

VAULT_PATH="${VAULT_PATH:-}"
DRY_RUN="${DRY_RUN:-true}"

if [ -z "$VAULT_PATH" ]; then
    echo "ERROR: Set VAULT_PATH environment variable"
    exit 1
fi

echo "=== Vault Migration Script ==="
echo "Vault path: $VAULT_PATH"
echo "Dry run: $DRY_RUN"
echo ""

# Counters
total_files=0
domain_fixed=0
scope_removed=0
tags_added=0
related_missing=0

# Temp file for processing
tmpfile=$(mktemp)
trap "rm -f $tmpfile" EXIT

# Process each markdown file
find "$VAULT_PATH" -name "*.md" -type f ! -path "*/.trash/*" 2>/dev/null | while read file; do
    total_files=$((total_files + 1))
    filename=$(basename "$file" .md)
    
    cp "$file" "$tmpfile"
    changed=false
    
    # Get current domain
    current_domain=$(grep "^domain:" "$file" | head -1 | sed 's/^domain:[ ]*//')
    
    # 1. Normalize domain
    if [ -n "$current_domain" ]; then
        new_domain=""
        case "$current_domain" in
            # AI domains
            ml-optimization|ml-training|ml-architecture|ml-pipeline) new_domain="ai-training" ;;
            ml-data|nlp) new_domain="ai-nlp" ;;
            neuro-symbolic-reasoning|ai-inference|query-pipeline|graph-rag) new_domain="ai-inference" ;;
            ai-orchestration|ai-agents|ai-frameworks|ai-tooling|agent-methodology|agentic-systems|biotech-ai-platform) new_domain="ai-agents" ;;
            personal-knowledge-profiling|social-media-aggregation|web-scraping) new_domain="ai-nlp" ;;
            
            # Dev tools domains
            developer-tools|developer-tooling|runtime|infrastructure|storage|pipeline|version-control|web-monitoring|data-storage) new_domain="dev-tools-cli" ;;
            testing|quality-assurance) new_domain="dev-tools-testing" ;;
            engineering-process|engineering-discipline|process|engineering-leadership|retrospective|group-coordination|architecture|communication) new_domain="dev-tools-workflow" ;;
            knowledge-management) new_domain="dev-tools-knowledge" ;;
            
            # Security domains
            security) new_domain="security-hardening" ;;
            security-infrastructure|agentic-security|agentic-security-gateway) new_domain="security-gateway" ;;
            regulatory-compliance) new_domain="security-compliance" ;;
            
            # Finance domains
            quantitative-finance|backtesting) new_domain="finance-quant" ;;
            fintech-loyalty) new_domain="finance-fintech" ;;
            
            # Frontend domains
            frontend|presentations) new_domain="frontend-ui" ;;
            
            # Calendar domains
            calendar-federation|calendar-coordination|calendar-sync) new_domain="calendar-sync" ;;
            
            # Already valid
            ai-training|ai-inference|ai-agents|ai-nlp|dev-tools-cli|dev-tools-testing|dev-tools-workflow|dev-tools-knowledge|security-gateway|security-hardening|security-compliance|finance-quant|finance-fintech|frontend-ui|frontend-performance|calendar-sync) new_domain="" ;;
            
            # Unknown
            *) echo "UNKNOWN: $filename has domain '$current_domain'" ;;
        esac
        
        if [ -n "$new_domain" ]; then
            echo "DOMAIN: $current_domain -> $new_domain ($filename)"
            sed -i '' "s/^domain:.*/domain: $new_domain/" "$tmpfile"
            domain_fixed=$((domain_fixed + 1))
            changed=true
            current_domain="$new_domain"
        fi
    fi
    
    # 2. Remove scope from frontmatter
    if grep -q "^scope:" "$tmpfile"; then
        echo "SCOPE: Removing from $filename"
        sed -i '' '/^scope:/d' "$tmpfile"
        scope_removed=$((scope_removed + 1))
        changed=true
    fi
    
    # 3. Add tags if missing
    if ! grep -q "^## Tags" "$tmpfile" && [ -n "$current_domain" ]; then
        tag=""
        case "$current_domain" in
            ai-training) tag="#ai/training" ;;
            ai-inference) tag="#ai/inference" ;;
            ai-agents) tag="#ai/agents" ;;
            ai-nlp) tag="#ai/nlp" ;;
            dev-tools-cli) tag="#dev-tools/cli" ;;
            dev-tools-testing) tag="#dev-tools/testing" ;;
            dev-tools-workflow) tag="#dev-tools/workflow" ;;
            dev-tools-knowledge) tag="#dev-tools/knowledge" ;;
            security-gateway) tag="#security/gateway" ;;
            security-hardening) tag="#security/hardening" ;;
            security-compliance) tag="#security/compliance" ;;
            finance-quant) tag="#finance/quant" ;;
            finance-fintech) tag="#finance/fintech" ;;
            frontend-ui) tag="#frontend/ui" ;;
            frontend-performance) tag="#frontend/performance" ;;
            calendar-sync) tag="#calendar/sync" ;;
        esac
        
        if [ -n "$tag" ]; then
            echo "TAGS: Adding $tag to $filename"
            echo "" >> "$tmpfile"
            echo "## Tags" >> "$tmpfile"
            echo "" >> "$tmpfile"
            echo "$tag" >> "$tmpfile"
            tags_added=$((tags_added + 1))
            changed=true
        fi
    fi
    
    # 4. Check for Related section
    if ! grep -q "^## Related" "$tmpfile"; then
        wikilink_count=$(grep -c "\[\[" "$tmpfile" 2>/dev/null || echo "0")
        if [ "$wikilink_count" -gt 0 ]; then
            echo "RELATED: Missing in $filename (has $wikilink_count wikilinks)"
            related_missing=$((related_missing + 1))
        fi
    fi
    
    # Write changes if not dry run
    if [ "$changed" = true ] && [ "$DRY_RUN" != "true" ]; then
        cp "$tmpfile" "$file"
    fi
done

echo ""
echo "=== Summary ==="
echo "Domains normalized: $domain_fixed"
echo "Scope removed: $scope_removed"  
echo "Tags added: $tags_added"
echo "Notes missing Related: $related_missing"
echo ""
if [ "$DRY_RUN" = "true" ]; then
    echo "DRY RUN - no changes made"
    echo "Run with DRY_RUN=false to apply changes"
fi
