// this file only exists to allow the Docker image to pre-build the cargo dependencies and will be replaced by the user's run.rs

pub fn run(input: String) -> Option<String> {
    
    let out = String::from(format!("hello {}", input));
    
    return Some(out);
}