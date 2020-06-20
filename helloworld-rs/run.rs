
#[no_mangle]
pub fn run(input: String) -> Option<String> {    
    Some(String::from(format!("Hello {}", input)))
}