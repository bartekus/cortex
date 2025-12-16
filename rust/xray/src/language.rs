use std::path::Path;

/// Detects language from file path (extension based).
/// Returns explicit "Unknown" if not matched, or the canonical language name.
pub fn detect_language(path: &Path) -> String {
    // Special filenames
    if let Some(name) = path.file_name().and_then(|s| s.to_str()) {
        if name.eq_ignore_ascii_case("Dockerfile") {
            return "Dockerfile".to_string();
        }
        if name.eq_ignore_ascii_case("Makefile") {
            return "Makefile".to_string();
        }
    }

    // Extensions
    if let Some(ext) = path.extension().and_then(|s| s.to_str()) {
        match ext.to_lowercase().as_str() {
            "go" => "Go",
            "rs" => "Rust",
            "md" => "Markdown",
            "json" => "JSON",
            "js" => "JavaScript",
            "ts" => "TypeScript",
            "yaml" | "yml" => "YAML",
            "toml" => "TOML",
            "sh" | "bash" => "Shell",
            "html" | "htm" => "HTML",
            "css" => "CSS",
            "sql" => "SQL",
            "py" => "Python",
            "java" => "Java",
            "c" | "h" => "C",
            "cpp" | "hpp" | "cc" | "cxx" => "C++",
            "tf" => "Terraform",
            "txt" | "text" => "Text",
            _ => "Unknown", // Or leave empty? Spec implies "languages" map. Unknowns usually ignored in stats? 
                            // let's return "Unknown" so it's explicit for now, but usually we might exclude from stats.
                            // The user said "others Unknown or skip (choose + lock)". I will lock to "Unknown".
        }.to_string()
    } else {
        "Unknown".to_string()
    }
}
