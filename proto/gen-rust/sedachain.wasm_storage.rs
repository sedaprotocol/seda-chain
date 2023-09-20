// @generated
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Wasm {
    #[prost(bytes="vec", tag="1")]
    pub hash: ::prost::alloc::vec::Vec<u8>,
    #[prost(bytes="vec", tag="2")]
    pub bytecode: ::prost::alloc::vec::Vec<u8>,
    #[prost(enumeration="WasmType", tag="3")]
    pub wasm_type: i32,
    #[prost(message, optional, tag="4")]
    pub added_at: ::core::option::Option<::prost_types::Timestamp>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Params {
    #[prost(uint64, tag="1")]
    pub max_wasm_size: u64,
}
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum WasmType {
    Unspecified = 0,
    DataRequest = 1,
    Tally = 2,
    DataRequestExecutor = 3,
    Relayer = 4,
}
impl WasmType {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            WasmType::Unspecified => "WASM_TYPE_UNSPECIFIED",
            WasmType::DataRequest => "WASM_TYPE_DATA_REQUEST",
            WasmType::Tally => "WASM_TYPE_TALLY",
            WasmType::DataRequestExecutor => "WASM_TYPE_DATA_REQUEST_EXECUTOR",
            WasmType::Relayer => "WASM_TYPE_RELAYER",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "WASM_TYPE_UNSPECIFIED" => Some(Self::Unspecified),
            "WASM_TYPE_DATA_REQUEST" => Some(Self::DataRequest),
            "WASM_TYPE_TALLY" => Some(Self::Tally),
            "WASM_TYPE_DATA_REQUEST_EXECUTOR" => Some(Self::DataRequestExecutor),
            "WASM_TYPE_RELAYER" => Some(Self::Relayer),
            _ => None,
        }
    }
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GenesisState {
    #[prost(message, optional, tag="1")]
    pub params: ::core::option::Option<Params>,
    #[prost(message, repeated, tag="2")]
    pub wasms: ::prost::alloc::vec::Vec<Wasm>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct QueryDataRequestWasmRequest {
    #[prost(string, tag="1")]
    pub hash: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct QueryDataRequestWasmResponse {
    #[prost(message, optional, tag="1")]
    pub wasm: ::core::option::Option<Wasm>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct QueryDataRequestWasmsRequest {
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct QueryDataRequestWasmsResponse {
    #[prost(string, repeated, tag="1")]
    pub hash_type_pairs: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct QueryOverlayWasmRequest {
    #[prost(string, tag="1")]
    pub hash: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct QueryOverlayWasmResponse {
    #[prost(message, optional, tag="1")]
    pub wasm: ::core::option::Option<Wasm>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct QueryOverlayWasmsRequest {
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct QueryOverlayWasmsResponse {
    #[prost(string, repeated, tag="1")]
    pub hash_type_pairs: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgStoreDataRequestWasm {
    #[prost(string, tag="1")]
    pub sender: ::prost::alloc::string::String,
    #[prost(bytes="vec", tag="2")]
    pub wasm: ::prost::alloc::vec::Vec<u8>,
    #[prost(enumeration="WasmType", tag="3")]
    pub wasm_type: i32,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgStoreDataRequestWasmResponse {
    #[prost(string, tag="1")]
    pub hash: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgStoreOverlayWasm {
    #[prost(string, tag="1")]
    pub sender: ::prost::alloc::string::String,
    #[prost(bytes="vec", tag="2")]
    pub wasm: ::prost::alloc::vec::Vec<u8>,
    #[prost(enumeration="WasmType", tag="3")]
    pub wasm_type: i32,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgStoreOverlayWasmResponse {
    #[prost(string, tag="1")]
    pub hash: ::prost::alloc::string::String,
}
include!("sedachain.wasm_storage.tonic.rs");
// @@protoc_insertion_point(module)