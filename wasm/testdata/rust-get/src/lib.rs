use suborbital::runnable;
use suborbital::cache;

struct RustGet{}

impl runnable::Runnable for RustGet {
    fn run(&self, _: Vec<u8>) -> Option<Vec<u8>> {
        let cache_val = cache::get("important").unwrap();
    
        Some(cache_val)
    }
}


// initialize the runner, do not edit below //
static RUNNABLE: &RustGet = &RustGet{};

#[no_mangle]
pub extern fn init() {
    runnable::set(RUNNABLE);
}
