/** Many messages are simpler to build and manage using native types that aren't 100% identical to what is expected by the canonical JSON representation of those messages.
 *
 * To deal with this, generated message classes use the best native types but provide functions to convert to the true wire format before transmission.
 * 
 * For example, protojson of uint64 is a string-encoded number, but the native type is still just a number.
 */
export interface ProtoJSONCompatible {
	/** Convert native fields to canonical protojson format
	 *
	 * e.g. 64-bit numbers as strings, bytes as base64, oneofs as specific instance fields
	 * */
	ToProtoJSON(): Promise<Object>;
}