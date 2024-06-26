// Generated by the protocol buffer compiler. DO NOT EDIT!
// source: VISSv2.proto

// Generated files should ignore deprecation warnings
@file:Suppress("DEPRECATION")
package grpcProtobufMessages;

@kotlin.jvm.JvmName("-initializesetRequestMessage")
public inline fun setRequestMessage(block: grpcProtobufMessages.SetRequestMessageKt.Dsl.() -> kotlin.Unit): grpcProtobufMessages.VISSv2OuterClass.SetRequestMessage =
  grpcProtobufMessages.SetRequestMessageKt.Dsl._create(grpcProtobufMessages.VISSv2OuterClass.SetRequestMessage.newBuilder()).apply { block() }._build()
/**
 * Protobuf type `grpcProtobufMessages.SetRequestMessage`
 */
public object SetRequestMessageKt {
  @kotlin.OptIn(com.google.protobuf.kotlin.OnlyForUseByGeneratedProtoCode::class)
  @com.google.protobuf.kotlin.ProtoDslMarker
  public class Dsl private constructor(
    private val _builder: grpcProtobufMessages.VISSv2OuterClass.SetRequestMessage.Builder
  ) {
    public companion object {
      @kotlin.jvm.JvmSynthetic
      @kotlin.PublishedApi
      internal fun _create(builder: grpcProtobufMessages.VISSv2OuterClass.SetRequestMessage.Builder): Dsl = Dsl(builder)
    }

    @kotlin.jvm.JvmSynthetic
    @kotlin.PublishedApi
    internal fun _build(): grpcProtobufMessages.VISSv2OuterClass.SetRequestMessage = _builder.build()

    /**
     * `string Path = 1;`
     */
    public var path: kotlin.String
      @JvmName("getPath")
      get() = _builder.getPath()
      @JvmName("setPath")
      set(value) {
        _builder.setPath(value)
      }
    /**
     * `string Path = 1;`
     */
    public fun clearPath() {
      _builder.clearPath()
    }

    /**
     * `string Value = 2;`
     */
    public var value: kotlin.String
      @JvmName("getValue")
      get() = _builder.getValue()
      @JvmName("setValue")
      set(value) {
        _builder.setValue(value)
      }
    /**
     * `string Value = 2;`
     */
    public fun clearValue() {
      _builder.clearValue()
    }

    /**
     * `optional string Authorization = 3;`
     */
    public var authorization: kotlin.String
      @JvmName("getAuthorization")
      get() = _builder.getAuthorization()
      @JvmName("setAuthorization")
      set(value) {
        _builder.setAuthorization(value)
      }
    /**
     * `optional string Authorization = 3;`
     */
    public fun clearAuthorization() {
      _builder.clearAuthorization()
    }
    /**
     * `optional string Authorization = 3;`
     * @return Whether the authorization field is set.
     */
    public fun hasAuthorization(): kotlin.Boolean {
      return _builder.hasAuthorization()
    }

    /**
     * `optional string RequestId = 4;`
     */
    public var requestId: kotlin.String
      @JvmName("getRequestId")
      get() = _builder.getRequestId()
      @JvmName("setRequestId")
      set(value) {
        _builder.setRequestId(value)
      }
    /**
     * `optional string RequestId = 4;`
     */
    public fun clearRequestId() {
      _builder.clearRequestId()
    }
    /**
     * `optional string RequestId = 4;`
     * @return Whether the requestId field is set.
     */
    public fun hasRequestId(): kotlin.Boolean {
      return _builder.hasRequestId()
    }
  }
}
public inline fun grpcProtobufMessages.VISSv2OuterClass.SetRequestMessage.copy(block: grpcProtobufMessages.SetRequestMessageKt.Dsl.() -> kotlin.Unit): grpcProtobufMessages.VISSv2OuterClass.SetRequestMessage =
  grpcProtobufMessages.SetRequestMessageKt.Dsl._create(this.toBuilder()).apply { block() }._build()

