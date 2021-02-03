/**
 * 
 * This file represents the Rust "API" for Hive WASM runnables. The functions defined herein are used to exchange data
 * between the host (Hive, written in Go) and the Runnable (a WASM module, in this case written in Rust).
 * 
 */

// a small wrapper to hold our dynamic Runnable
struct State <'a> {
    ident: i32,
    runnable: &'a dyn runnable::Runnable
}

// something to hold down the fort until a real Runnable is set
struct DefaultRunnable {}
impl runnable::Runnable for DefaultRunnable {
    fn run(&self, _input: Vec<u8>) -> Option<Vec<u8>> {
        return None;
    }
}

// the state that holds the user-provided Runnable and the current ident
static mut STATE: State = State {
    ident: 0,
    runnable: &DefaultRunnable{},
};

pub mod runnable {
    use std::mem;
    use std::slice;

    extern {
        fn return_result(result_pointer: *const u8, result_size: i32, ident: i32);
    }

    pub trait Runnable {
        fn run(&self, input: Vec<u8>) -> Option<Vec<u8>>;
    }

    pub fn set(runnable: &'static dyn Runnable) {
        unsafe {
            super::STATE.runnable = runnable;
        }
    }
    
    #[no_mangle]
    pub extern fn allocate(size: i32) -> *const u8 {
        let mut buffer = Vec::with_capacity(size as usize);
        let buffer_slice = buffer.as_mut_slice();
        let pointer = buffer_slice.as_mut_ptr();
        mem::forget(buffer_slice);
    
        pointer as *const u8
    }
    
    #[no_mangle]
    pub extern fn deallocate(pointer: *const u8, size: i32) {
        unsafe {
            let _ = slice::from_raw_parts(pointer, size as usize);
        }
    }
    
    #[no_mangle]
    pub extern fn run_e(pointer: *const u8, size: i32, ident: i32) {
        unsafe { super::STATE.ident = ident };
    
        // rebuild the memory into something usable
        let in_slice: &[u8] = unsafe { 
            slice::from_raw_parts(pointer, size as usize) 
        };
    
        let in_bytes = Vec::from(in_slice);
    
        // call the runnable and check its result
        let result: Vec<u8> = unsafe { match super::STATE.runnable.run(in_bytes) {
            Some(val) => val,
            None => Vec::from("run returned no data"), 
        } };
    
        let result_slice = result.as_slice();
        let result_size = result_slice.len();
    
    
        // call back to hive to return the result
        unsafe { 
            return_result(result_slice.as_ptr() as *const u8, result_size as i32, ident); 
        }
    }
}

pub mod http {
    use std::collections::BTreeMap;
	use std::slice;

    static METHOD_GET: i32 = 1;
    static METHOD_POST: i32 = 2;
    static METHOD_PATCH: i32 = 3;
    static METHOD_DELETE: i32 = 4;

    extern {
        fn fetch_url(method: i32, url_pointer: *const u8, url_size: i32, body_pointer: *const u8, body_size: i32, dest_pointer: *const u8, dest_max_size: i32, ident: i32) -> i32;
    }

    pub fn get(url: &str, headers: Option<BTreeMap<&str, &str>>) -> Vec<u8> {
		return do_request(METHOD_GET, url, None, headers);
	}
    
    pub fn post(url: &str, body: Option<Vec<u8>>, headers: Option<BTreeMap<&str, &str>>) -> Vec<u8> {
		return do_request(METHOD_POST, url, body, headers);
	}
    
    pub fn patch(url: &str, body: Option<Vec<u8>>, headers: Option<BTreeMap<&str, &str>>) -> Vec<u8> {
		return do_request(METHOD_PATCH, url, body, headers);
	}
    
    pub fn delete(url: &str, headers: Option<BTreeMap<&str, &str>>) -> Vec<u8> {
		return do_request(METHOD_DELETE, url, None, headers);
	}

	fn do_request(method: i32, url: &str, body: Option<Vec<u8>>, headers: Option<BTreeMap<&str, &str>>) -> Vec<u8> {
        // the URL gets encoded with headers added on the end, seperated by ::
	    // eg. https://google.com/somepage::authorization:bearer qdouwrnvgoquwnrg::anotherheader:nicetomeetyou
        let header_string = render_header_string(headers);
        
        let url_string = match header_string {
            Some(h) => format!("{}::{}", url, h),
            None => String::from(url)
        };
        
        let mut dest_pointer: *const u8;
        let mut dest_size: i32;
        let mut capacity: i32 = 256000;

        let body_pointer: *const u8;
        let mut body_size: i32 = 0;

        match body {
            Some(b) => {
                let body_slice = b.as_slice();
                body_pointer = body_slice.as_ptr();
                body_size = b.len() as i32;
            },
            None => body_pointer = 0 as *const u8
        }
        
        // make the request, and if the response size is greater than that of capacity, increase the capacity and try again
        loop {
            let cap = &mut capacity;
            
            let mut dest_bytes = Vec::with_capacity(*cap as usize);
            let dest_slice = dest_bytes.as_mut_slice();
            dest_pointer = dest_slice.as_mut_ptr() as *const u8;
            
            // do the request over FFI
            dest_size = unsafe { fetch_url(method, url_string.as_str().as_ptr(), url_string.len() as i32, body_pointer, body_size, dest_pointer, *cap, super::STATE.ident) };

            if dest_size < 0 {
                return Vec::from(format!("request_failed:{}", dest_size))
            } else if dest_size > *cap {
                *cap = dest_size;
            } else {
                break;
            }
        }

        let result: &[u8] = unsafe {
            slice::from_raw_parts(dest_pointer, dest_size as usize)
        };

        return Vec::from(result)
	}
	
	fn render_header_string(headers: Option<BTreeMap<&str, &str>>) -> Option<String> {
        let mut rendered: String = String::from("");
        
        match headers {
            Some(h) => {
                for key in h.keys() {
                    rendered.push_str(key);
                    rendered.push_str(":");
        
                    let val: &str = match h.get(key) {
                        Some(v) => v,
                        None => "",
                    };
        
                    rendered.push_str(val);
                    rendered.push_str("::")
                }
            },
            None => return None,
        }

		return Some(String::from(rendered.trim_end_matches("::")));
	}
}

pub mod cache {
    use std::slice;

    extern {
        fn cache_set(key_pointer: *const u8, key_size: i32, value_pointer: *const u8, value_size: i32, ttl: i32, ident: i32) -> i32;
        fn cache_get(key_pointer: *const u8, key_size: i32, dest_pointer: *const u8, dest_max_size: i32, ident: i32) -> i32;
    }

    pub fn set(key: &str, val: Vec<u8>, ttl: i32) {
        let val_slice = val.as_slice();
        let val_ptr = val_slice.as_ptr();

        unsafe {
            cache_set(key.as_ptr(), key.len() as i32, val_ptr, val.len() as i32, ttl, super::STATE.ident);
        }
    }

    pub fn get(key: &str) -> Option<Vec<u8>> {
        let mut dest_pointer: *const u8;
        let mut result_size: i32;
        let mut capacity: i32 = 1024;

        // make the request, and if the response size is greater than that of capacity, increase the capacity and try again
        loop {
            let cap = &mut capacity;

            let mut dest_bytes = Vec::with_capacity(*cap as usize);
            let dest_slice = dest_bytes.as_mut_slice();
            dest_pointer = dest_slice.as_mut_ptr() as *const u8;
    
            // do the request over FFI
            result_size = unsafe { cache_get(key.as_ptr(), key.len() as i32, dest_pointer, *cap, super::STATE.ident) };

            if result_size < 0 {
                return None;
            } else if result_size > *cap {
                super::log::info(format!("increasing capacity, need {}", result_size).as_str());
                *cap = result_size;
            } else {
                break;
            }
        }

        let result: &[u8] = unsafe {
            slice::from_raw_parts(dest_pointer, result_size as usize)
        };

        Some(Vec::from(result))
    }
}

pub mod req {
    use std::slice;
    use super::util;

    extern {
        fn request_get_field(field_type: i32, key_pointer: *const u8, key_size: i32, dest_pointer: *const u8, dest_max_size: i32, ident: i32) -> i32;
    }

    static FIELD_TYPE_META: i32 = 0 as i32;
    static FIELD_TYPE_BODY: i32 = 1 as i32;
    static FIELD_TYPE_HEADER: i32 = 2 as i32;
    static FIELD_TYPE_PARAMS: i32 = 3 as i32;
    static FIELD_TYPE_STATE: i32 = 4 as i32;

    pub fn method() -> String {
        match get_field(FIELD_TYPE_META, "method") {
            Some(bytes) => return util::to_string(bytes),
            None => return String::from("")
        }
    }
    
    pub fn url() -> String {
        match get_field(FIELD_TYPE_META, "url") {
            Some(bytes) => return util::to_string(bytes),
            None => return String::from("")
        }
    }
    
    pub fn id() -> String {
        match get_field(FIELD_TYPE_META, "id") {
            Some(bytes) => return util::to_string(bytes),
            None => return String::from("")
        }
    }
    
    pub fn body_raw() -> Vec<u8> {
        match get_field(FIELD_TYPE_META, "body") {
            Some(bytes) => return bytes,
            None => return util::to_vec(String::from(""))
        }
    }

    pub fn body_field(key: &str) -> String {
        match get_field(FIELD_TYPE_BODY, key) {
            Some(bytes) => return util::to_string(bytes),
            None => return String::from("")
        }
    }
    
    pub fn header(key: &str) -> String {
        match get_field(FIELD_TYPE_HEADER, key) {
            Some(bytes) => return util::to_string(bytes),
            None => return String::from("")
        }
    }
    
    pub fn url_param(key: &str) -> String {
        match get_field(FIELD_TYPE_PARAMS, key) {
            Some(bytes) => return util::to_string(bytes),
            None => return String::from("")
        }
    }
    
    pub fn state(key: &str) -> String {
        match get_field(FIELD_TYPE_STATE, key) {
            Some(bytes) => return util::to_string(bytes),
            None => return String::from("")
        }
    }
    
    fn get_field(field_type: i32, key: &str) -> Option<Vec<u8>> {
        let mut dest_pointer: *const u8;
        let mut result_size: i32;
        let mut capacity: i32 = 1024;

        // make the request, and if the response size is greater than that of capacity, increase the capacity and try again
        loop {
            let cap = &mut capacity;

            let mut dest_bytes = Vec::with_capacity(*cap as usize);
            let dest_slice = dest_bytes.as_mut_slice();
            dest_pointer = dest_slice.as_mut_ptr() as *const u8;
    
            // do the request over FFI
            result_size = unsafe { request_get_field(field_type, key.as_ptr(), key.len() as i32, dest_pointer, *cap, super::STATE.ident) };

            if result_size < 0 {
                return None;
            } else if result_size > *cap {
                super::log::info(format!("increasing capacity, need {}", result_size).as_str());
                *cap = result_size;
            } else {
                break;
            }
        }

        let result: &[u8] = unsafe {
            slice::from_raw_parts(dest_pointer, result_size as usize)
        };

        Some(Vec::from(result))
    }
}

pub mod log {
    extern {
        fn log_msg(pointer: *const u8, result_size: i32, level: i32, ident: i32);
    }

    pub fn info(msg: &str) {
        log_at_level(msg, 3)
    }
    
    pub fn warn(msg: &str) {
        log_at_level(msg, 2)
    }
    
    pub fn error(msg: &str) {
        log_at_level(msg, 1)
    }

    fn log_at_level(msg: &str, level: i32) {
        let msg_vec = Vec::from(msg);
        let msg_slice = msg_vec.as_slice();
        let pointer = msg_slice.as_ptr();

        unsafe { log_msg(pointer, msg_slice.len() as i32, level, super::STATE.ident) };
    }
}

pub mod util {
    pub fn to_string(input: Vec<u8>) -> String {
        String::from_utf8(input).unwrap()
    }

    pub fn to_vec(input: String) -> Vec<u8> {
        input.as_bytes().to_vec()
    }

    pub fn str_to_vec(input: &str) -> Vec<u8> {
        String::from(input).as_bytes().to_vec()
    }
}