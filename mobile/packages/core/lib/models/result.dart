/// A discriminated union representing either a successful value or a failure.
///
/// Use pattern matching (switch expressions) to handle both cases:
/// ```dart
/// final result = await repo.getJob(id);
/// switch (result) {
///   case Success(:final data):
///     print('Got job: ${data.title}');
///   case Failure(:final message):
///     print('Error: $message');
/// }
/// ```
sealed class Result<T> {
  const Result();

  /// Whether this result is a success.
  bool get isSuccess => this is Success<T>;

  /// Whether this result is a failure.
  bool get isFailure => this is Failure<T>;

  /// Get the data if this is a success, otherwise null.
  T? get dataOrNull {
    final self = this;
    if (self is Success<T>) return self.data;
    return null;
  }

  /// Get the error message if this is a failure, otherwise null.
  String? get errorOrNull {
    final self = this;
    if (self is Failure<T>) return self.message;
    return null;
  }

  /// Transform the success value, leaving failures unchanged.
  Result<R> map<R>(R Function(T data) transform) {
    final self = this;
    if (self is Success<T>) {
      return Success(transform(self.data));
    }
    final failure = self as Failure<T>;
    return Failure(failure.message, statusCode: failure.statusCode);
  }
}

/// A successful result containing [data].
class Success<T> extends Result<T> {
  final T data;
  const Success(this.data);

  @override
  String toString() => 'Success($data)';
}

/// A failed result containing an error [message] and optional HTTP
/// [statusCode].
class Failure<T> extends Result<T> {
  final String message;
  final int? statusCode;
  const Failure(this.message, {this.statusCode});

  @override
  String toString() => 'Failure($message, statusCode: $statusCode)';
}
