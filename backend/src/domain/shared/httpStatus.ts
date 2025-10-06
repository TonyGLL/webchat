/**
 * Hyperyexy Transfer Protocol (HTTP) response status codes.
 * @see {@link https://en.wikipedia.org/wiki/List_of_HTTP_status_codes}
 */
export enum HttpStatusCode {
    // Informational 1xx
    /**
     * The server has received the request headers and the client should proceed to send the request body.
     */
    CONTINUE = 100,

    /**
     * The requester has asked the server to switch protocols.
     */
    SWITCHING_PROTOCOLS = 101,

    /**
     * Used to return some response headers before final HTTP message.
     */
    PROCESSING = 102,

    /**
     * Used to indicate to the client that the server is likely to send a final response.
     */
    EARLY_HINTS = 103,

    // Success 2xx
    /**
     * Standard response for successful HTTP requests.
     */
    OK = 200,

    /**
     * The request has been fulfilled, resulting in the creation of a new resource.
     */
    CREATED = 201,

    /**
     * The request has been accepted for processing, but processing has not been completed.
     */
    ACCEPTED = 202,

    /**
     * The server is a transforming proxy that received a 200 OK from its origin.
     */
    NON_AUTHORITATIVE_INFORMATION = 203,

    /**
     * The server successfully processed the request and is not returning any content.
     */
    NO_CONTENT = 204,

    /**
     * The server successfully processed the request, but is not returning any content.
     */
    RESET_CONTENT = 205,

    /**
     * The server is delivering only part of the resource due to a range header.
     */
    PARTIAL_CONTENT = 206,

    /**
     * The message body that follows is an XML message and can contain multiple response codes.
     */
    MULTI_STATUS = 207,

    /**
     * The members of a DAV binding have already been enumerated.
     */
    ALREADY_REPORTED = 208,

    /**
     * The server has fulfilled a GET request for the resource.
     */
    IM_USED = 226,

    // Redirection 3xx
    /**
     * Indicates multiple options for the resource.
     */
    MULTIPLE_CHOICES = 300,

    /**
     * This and all future requests should be directed to the given URI.
     */
    MOVED_PERMANENTLY = 301,

    /**
     * Temporary redirect for HTTP/1.0 clients.
     */
    FOUND = 302,

    /**
     * The response to the request can be found under another URI using a GET method.
     */
    SEE_OTHER = 303,

    /**
     * Indicates that the resource has not been modified since last request.
     */
    NOT_MODIFIED = 304,

    /**
     * Deprecated. The requested resource is available only through a proxy.
     */
    USE_PROXY = 305,

    /**
     * No longer used. Reserved for future use.
     */
    UNUSED = 306,

    /**
     * Temporary redirect for HTTP/1.1 clients.
     */
    TEMPORARY_REDIRECT = 307,

    /**
     * Permanent redirect for HTTP/1.1 clients.
     */
    PERMANENT_REDIRECT = 308,

    // Client Errors 4xx
    /**
     * The server cannot process the request due to an apparent client error.
     */
    BAD_REQUEST = 400,

    /**
     * Authentication is required and has failed or not been provided.
     */
    UNAUTHORIZED = 401,

    /**
     * Reserved for future use. Original meaning: "Payment required".
     */
    PAYMENT_REQUIRED = 402,

    /**
     * The request was valid, but the server is refusing action.
     */
    FORBIDDEN = 403,

    /**
     * The requested resource could not be found.
     */
    NOT_FOUND = 404,

    /**
     * A request method is not supported for the requested resource.
     */
    METHOD_NOT_ALLOWED = 405,

    /**
     * The requested resource is capable of generating only content not acceptable.
     */
    NOT_ACCEPTABLE = 406,

    /**
     * The client must first authenticate itself with the proxy.
     */
    PROXY_AUTHENTICATION_REQUIRED = 407,

    /**
     * The server timed out waiting for the request.
     */
    REQUEST_TIMEOUT = 408,

    /**
     * The request could not be processed because of conflict.
     */
    CONFLICT = 409,

    /**
     * The resource is no longer available and will not be available again.
     */
    GONE = 410,

    /**
     * The request did not specify the length of its content.
     */
    LENGTH_REQUIRED = 411,

    /**
     * The server does not meet one of the preconditions.
     */
    PRECONDITION_FAILED = 412,

    /**
     * The request is larger than the server is willing or able to process.
     */
    PAYLOAD_TOO_LARGE = 413,

    /**
     * The URI provided was too long for the server to process.
     */
    URI_TOO_LONG = 414,

    /**
     * The request entity has a media type which the server does not support.
     */
    UNSUPPORTED_MEDIA_TYPE = 415,

    /**
     * The client has asked for a portion of the file, but the server cannot supply that portion.
     */
    RANGE_NOT_SATISFIABLE = 416,

    /**
     * The server cannot meet the requirements of the Expect request-header field.
     */
    EXPECTATION_FAILED = 417,

    /**
     * I'm a teapot  = RFC2324).
     */
    I_AM_A_TEAPOT = 418,

    /**
     * The request was directed at a server that is not able to produce a response.
     */
    MISDIRECTED_REQUEST = 421,

    /**
     * The request was well-formed but unable to be followed due to semantic errors.
     */
    UNPROCESSABLE_ENTITY = 422,

    /**
     * The resource that is being accessed is locked.
     */
    LOCKED = 423,

    /**
     * The request failed due to failure of a previous request.
     */
    FAILED_DEPENDENCY = 424,

    /**
     * The server is unwilling to risk processing a request that might be replayed.
     */
    TOO_EARLY = 425,

    /**
     * The client should switch to a different protocol.
     */
    UPGRADE_REQUIRED = 426,

    /**
     * The origin server requires the request to be conditional.
     */
    PRECONDITION_REQUIRED = 428,

    /**
     * The user has sent too many requests in a given amount of time.
     */
    TOO_MANY_REQUESTS = 429,

    /**
     * The server is unwilling to process the request because header fields are too large.
     */
    REQUEST_HEADER_FIELDS_TOO_LARGE = 431,

    /**
     * A server operator has received a legal demand to deny access to a resource.
     */
    UNAVAILABLE_FOR_LEGAL_REASONS = 451,

    // Server Errors 5xx
    /**
     * A generic error message when an unexpected condition was encountered.
     */
    INTERNAL_SERVER_ERROR = 500,

    /**
     * The server either does not recognize the request method, or lacks the ability to fulfill.
     */
    NOT_IMPLEMENTED = 501,

    /**
     * The server was acting as a gateway or proxy and received an invalid response.
     */
    BAD_GATEWAY = 502,

    /**
     * The server is currently unavailable.
     */
    SERVICE_UNAVAILABLE = 503,

    /**
     * The server was acting as a gateway or proxy and did not receive a timely response.
     */
    GATEWAY_TIMEOUT = 504,

    /**
     * The server does not support the HTTP protocol version used in the request.
     */
    HTTP_VERSION_NOT_SUPPORTED = 505,

    /**
     * Transparent content negotiation for the request results in a circular reference.
     */
    VARIANT_ALSO_NEGOTIATES = 506,

    /**
     * The server is unable to store the representation needed to complete the request.
     */
    INSUFFICIENT_STORAGE = 507,

    /**
     * The server detected an infinite loop while processing the request.
     */
    LOOP_DETECTED = 508,

    /**
     * Further extensions to the request are required for the server to fulfill it.
     */
    NOT_EXTENDED = 510,

    /**
     * The client needs to authenticate to gain network access.
     */
    NETWORK_AUTHENTICATION_REQUIRED = 511
}