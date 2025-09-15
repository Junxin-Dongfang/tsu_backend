package contextkeys

type contextKey string

const TraceIDKey = contextKey("trace_id")
const UserIDKey = contextKey("user_id")
