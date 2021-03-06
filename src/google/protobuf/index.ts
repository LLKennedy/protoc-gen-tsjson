/**
 * This file is for special types exported by google protobufs, such as Any, Empty, Timestamp and Duration.
 * They get special treatment by the canonical JSON marshalling rules, so we give them special treatment here too.
 */

export class Any {
	constructor(data?: any) {
		this.value = data;
	}
	public value?: any;
	public ToProtoJSON(): any {
		return this.value;
	}
	public static async Parse(data: any): Promise<Any> {
		return new Any(data);
	}
}

export class Timestamp {
	constructor(ts?: Date) {
		this.timestamp = ts;
	}
	public timestamp?: Date;
	public ToProtoJSON(): string | undefined {
		return this.timestamp?.toISOString()
	}
	public static async Parse(data: any): Promise<Timestamp> {
		switch (typeof data) {
			case "object":
				if (!(data instanceof Date)) {
					// TODO: handle marshalling of datetime struct
					throw new Error("Non-date objects not supported for date parsing");
				}
				return new Timestamp(data);
			case "string":
			case "number":
				return new Timestamp(new Date(data));
			default:
				throw new Error("date can only be marshalled from string or number")
		}
	}
}

export class Duration {
	constructor(seconds?: number) {
		this.durationSeconds = seconds;
	}
	public durationSeconds?: number;
	public ToProtoJSON(): string {
		return (this.durationSeconds?.toFixed(9) ?? "0") + "s";
	}
	public static async Parse(data: any): Promise<Duration> {
		if (typeof data !== "string") {
			throw new Error("duration must be a string");
		}
		if (!data.endsWith("s")) {
			throw new Error("duration must end with s");
		}
		if (data.indexOf("s") !== (data.length - 1)) {
			throw new Error("duration must only contain one s")
		}
		data = data.replace("s", "");
		return new Duration(Number(data));
	}
}

export class Struct {
	constructor(data?: Object) {
		this.data = data;
	}
	public data?: Object;
	public ToProtoJSON(): Object | undefined {
		return this.data;
	}
	public static async Parse(data: any): Promise<Struct> {
		switch (typeof data) {
			case "object":
				return new Struct(data);
			case "string":
				return new Struct(JSON.parse(data));
			default:
				throw new Error("unimplemented");
		}
	}
}

export class Wrapper {
	public ToProtoJSON(): any {
		throw new Error("unimplemented");
	}
	public static async Parse(_: any): Promise<Wrapper> {
		throw new Error("unimplemented");
	}
}

export class FieldMask {
	public ToProtoJSON(): any {
		throw new Error("unimplemented");
	}
	public static async Parse(_: any): Promise<FieldMask> {
		throw new Error("unimplemented");
	}
}

export class ListValue {
	constructor(data?: any[]) {
		this.list = data;
	}
	public list?: any[];
	public ToProtoJSON(): any {
		throw new Error("unimplemented");
	}
	public static async Parse(_: any): Promise<ListValue> {
		throw new Error("unimplemented");
	}
}

export class Value {
	constructor(data?: any) {
		this.value = data;
	}
	public value?: any;
	public ToProtoJSON(): any {
		return this.value;
	}
	public static async Parse(data: any): Promise<Value> {
		return data;
	}
}

export class NullValue {
	public ToProtoJSON(): null {
		return null;
	}
	public static async Parse(_: any): Promise<null> {
		return null;
	}
}

export class Empty {
	public ToProtoJSON(): {} {
		return {};
	}
	public static async Parse(_: any): Promise<Empty> {
		return new Empty();
	}
}