
@_silgen_name("return_result_swift")
func return_result(result_pointer: UnsafeRawPointer, result_size: Int32, ident: Int32)
@_silgen_name("log_msg_swift")
func log_msg_swift(pointer: UnsafeRawPointer, size: Int32, level: Int32, ident: Int32)
@_silgen_name("cache_set_swift")
func cache_set_swift(key_pointer: UnsafeRawPointer, key_size: Int32, value_pointer: UnsafeRawPointer, value_size: Int32, ttl: Int32, ident: Int32) -> Int32
@_silgen_name("cache_get_swift")
func cache_get_swift(key_pointer: UnsafeRawPointer, key_size: Int32, dest_pointer: UnsafeRawPointer, dest_max_size: Int32, ident: Int32) -> Int32
@_silgen_name("request_get_field_swift")
func request_get_field(field_type: Int32, key_pointer: UnsafeRawPointer, key_size: Int32, dest_pointer: UnsafeRawPointer, dest_max_size: Int32, ident: Int32) -> Int32

// keep track of the current ident
var CURRENT_IDENT: Int32 = 0

// the Runnable instance currently being used
var RUNNABLE: Runnable = defaultRunnable()

// the protocol that users conform to to make their package a Runnable
public protocol Runnable {
    func run(input: String) -> String
}

// something to hold the Runnable's place until set is called
class defaultRunnable: Runnable {
    func run(input: String) -> String {
        return ""
    }
}

public func Set(runnable: Runnable) {
    RUNNABLE = runnable
}

public func CacheSet(key: String, value: String, ttl: Int) {    
    let keyFFI = toFFI(val: key)
    let valFFI = toFFI(val: value)

    let _ = cache_set_swift(key_pointer: keyFFI.0, key_size: keyFFI.1, value_pointer: valFFI.0, value_size: valFFI.1, ttl: Int32(ttl), ident: CURRENT_IDENT)
}

public func CacheGet(key: String) -> String {    
    var maxSize: Int32 = 256000
    var retVal = ""

    let keyFFI = toFFI(val: key)

    // loop until the returned size is within the defined max size, increasing it as needed
    while true {
        let ptr = allocate(size: Int32(maxSize))

        let resultSize = cache_get_swift(key_pointer: keyFFI.0, key_size: keyFFI.1, dest_pointer: ptr, dest_max_size: maxSize, ident: CURRENT_IDENT)

        if resultSize < 0 {
            retVal = "failed to get from cache"
            break
        } else if resultSize > maxSize {
            maxSize = resultSize
        } else {
            retVal = fromFFI(ptr: ptr, size: resultSize)
            break
        }
    }
    
    return retVal
}

public func LogInfo(msg: String) {
    log(msg: msg, level: 3)
}

public func LogWarn(msg: String) {
    log(msg: msg, level: 2)
}

public func LogErr(msg: String) {
    log(msg: msg, level: 1)
}

func log(msg: String, level: Int32) {
    let msgFFI = toFFI(val: msg)

    log_msg_swift(pointer: msgFFI.0, size: msgFFI.1, level: level, ident: CURRENT_IDENT)
}

let fieldTypeMeta = Int32(0)
let fieldTypeBody = Int32(1)
let fieldTypeHeader = Int32(2)
let fieldTypeParams = Int32(3)
let fieldTypeState = Int32(4)

public func ReqMethod() -> String {
    return requestGetField(fieldType: fieldTypeMeta, key: "method")
}

public func ReqURL() -> String {
    return requestGetField(fieldType: fieldTypeMeta, key: "url")
}

public func ReqID() -> String {
    return requestGetField(fieldType: fieldTypeMeta, key: "id")
}

public func ReqBodyRaw() -> String {
    return requestGetField(fieldType: fieldTypeMeta, key: "body")
}

public func ReqBodyField(key: String) -> String {
    return requestGetField(fieldType: fieldTypeBody, key: key)
}

public func ReqHeader(key: String) -> String {
    return requestGetField(fieldType: fieldTypeHeader, key: key)
}

public func ReqParam(key: String) -> String {
    return requestGetField(fieldType: fieldTypeParams, key: key)
}

public func State(key: String) -> String {
    return requestGetField(fieldType: fieldTypeState, key: key)
}

func requestGetField(fieldType: Int32, key: String) -> String {    
    var maxSize: Int32 = 1024
    var retVal = ""

    let keyFFI = toFFI(val: key)

    // loop until the returned size is within the defined max size, increasing it as needed
    while true {
        let ptr = allocate(size: Int32(maxSize))

        let resultSize = request_get_field(field_type: fieldType, key_pointer: keyFFI.0, key_size: keyFFI.1, dest_pointer: ptr, dest_max_size: maxSize, ident: CURRENT_IDENT)

        if resultSize < 0 {
            retVal = "failed to get request field"
            break
        } else if resultSize > maxSize {
            maxSize = resultSize
        } else {
            retVal = fromFFI(ptr: ptr, size: resultSize)
            break
        }
    }
    
    return retVal
}

@_cdecl("run_e")
func run_e(pointer: UnsafeRawPointer, size: Int32, ident: Int32) {
    CURRENT_IDENT = ident
    
    let inString = fromFFI(ptr: pointer, size: size)
    
    // call the user-provided run function
    let retString = RUNNABLE.run(input: inString)

    // convert the output to a usable pointer/size combo
    let retVal = toFFI(val: retString)

    return_result(result_pointer: retVal.0, result_size: retVal.1, ident: ident)
}

@_cdecl("allocate")
func allocate(size: Int32) -> UnsafeMutableRawPointer {
  return UnsafeMutableRawPointer.allocate(byteCount: Int(size), alignment: MemoryLayout<UInt8>.alignment)
}

@_cdecl("deallocate")
func deallocate(pointer: UnsafeRawPointer, size: Int32) {
    let ptr: UnsafePointer<UInt8> = pointer.bindMemory(to: UInt8.self, capacity: Int(size))
    ptr.deallocate()
}

func toFFI(val: String) -> (UnsafePointer<Int8>, Int32) {
    // create a nil (optional) pointer
    var ptr: UnsafePointer<Int8>? = UnsafePointer<Int8>(bitPattern: 0)
    let size = Int32(val.utf8.count)

    // grab the pointer in a closure and give the optional a real value
    let _ = val.withCString({ (valPtr) -> UInt in
        ptr = valPtr
        return 0
    })

    // unwrap the optional before returning
    return (ptr!, size)
}

func fromFFI(ptr: UnsafeRawPointer, size: Int32) -> String {
    let typed: UnsafePointer<UInt8> = ptr.bindMemory(to: UInt8.self, capacity: Int(size))
    let val = String(cString: typed)
    
    return val
}