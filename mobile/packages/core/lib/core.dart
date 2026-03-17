/// Core business logic, models, and services for the Seva platform.
///
/// This package contains shared data models, API client configuration,
/// repository abstractions, and platform services used across both
/// the customer and provider mobile applications.
library seva_core;

// API
export 'api/api_client.dart';

// Models
export 'models/user.dart';
export 'models/job.dart';
export 'models/provider.dart';
export 'models/category.dart';
export 'models/review.dart';
export 'models/notification.dart';

// Repositories
export 'repositories/auth_repository.dart';
export 'repositories/job_repository.dart';
export 'repositories/provider_repository.dart';
export 'repositories/notification_repository.dart';

// Services
export 'services/auth_service.dart';
export 'services/location_service.dart';
export 'services/storage_service.dart';
