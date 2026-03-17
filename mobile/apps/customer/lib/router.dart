import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:seva_core/core.dart';
import 'package:seva_ui_kit/ui_kit.dart';
import 'main.dart';
import 'screens/home/home_screen.dart';
import 'screens/search/search_screen.dart';
import 'screens/search/provider_detail_screen.dart';
import 'screens/booking/create_job_screen.dart';
import 'screens/booking/job_detail_screen.dart';
import 'screens/auth/login_screen.dart';
import 'screens/auth/register_screen.dart';
import 'screens/profile/profile_screen.dart';
import 'screens/notifications/notifications_screen.dart';
import 'screens/messages/messages_screen.dart';
import 'screens/messages/conversation_screen.dart';
import 'screens/map/map_screen.dart';

final routerProvider = Provider<GoRouter>((ref) {
  final authService = ref.watch(authServiceProvider);

  return GoRouter(
    initialLocation: '/',
    redirect: (context, state) {
      final isAuthenticated = authService.isAuthenticated;
      final isAuthRoute = state.matchedLocation.startsWith('/auth');

      if (!isAuthenticated && !isAuthRoute) {
        return '/auth/login';
      }
      if (isAuthenticated && isAuthRoute) {
        return '/';
      }
      return null;
    },
    routes: [
      // Main shell with bottom navigation
      ShellRoute(
        builder: (context, state, child) {
          return _CustomerShell(child: child);
        },
        routes: [
          GoRoute(
            path: '/',
            name: 'home',
            builder: (context, state) => const HomeScreen(),
          ),
          GoRoute(
            path: '/search',
            name: 'search',
            builder: (context, state) {
              final query = state.uri.queryParameters['q'];
              final categoryId = state.uri.queryParameters['category'];
              return SearchScreen(
                initialQuery: query,
                categoryId: categoryId,
              );
            },
          ),
          GoRoute(
            path: '/messages',
            name: 'messages',
            builder: (context, state) => const MessagesScreen(),
          ),
          GoRoute(
            path: '/map',
            name: 'map',
            builder: (context, state) => const MapScreen(),
          ),
          GoRoute(
            path: '/notifications',
            name: 'notifications',
            builder: (context, state) => const NotificationsScreen(),
          ),
          GoRoute(
            path: '/profile',
            name: 'profile',
            builder: (context, state) => const ProfileScreen(),
          ),
        ],
      ),

      // Detail routes (no bottom nav)
      GoRoute(
        path: '/provider/:id',
        name: 'provider-detail',
        builder: (context, state) {
          return ProviderDetailScreen(
            providerId: state.pathParameters['id']!,
          );
        },
      ),
      GoRoute(
        path: '/job/create',
        name: 'create-job',
        builder: (context, state) {
          final categoryId = state.uri.queryParameters['category'];
          final providerId = state.uri.queryParameters['provider'];
          return CreateJobScreen(
            categoryId: categoryId,
            providerId: providerId,
          );
        },
      ),
      GoRoute(
        path: '/job/:id',
        name: 'job-detail',
        builder: (context, state) {
          return JobDetailScreen(jobId: state.pathParameters['id']!);
        },
      ),
      GoRoute(
        path: '/conversation/:id',
        name: 'conversation',
        builder: (context, state) {
          final conversation = state.extra as Conversation?;
          return ConversationScreen(
            conversationId: state.pathParameters['id']!,
            conversation: conversation,
          );
        },
      ),

      // Auth routes
      GoRoute(
        path: '/auth/login',
        name: 'login',
        builder: (context, state) => const LoginScreen(),
      ),
      GoRoute(
        path: '/auth/register',
        name: 'register',
        builder: (context, state) => const RegisterScreen(),
      ),
    ],
  );
});

/// Shell widget providing bottom navigation for the customer app.
class _CustomerShell extends StatelessWidget {
  final Widget child;

  const _CustomerShell({required this.child});

  int _currentIndex(BuildContext context) {
    final location = GoRouterState.of(context).matchedLocation;
    if (location == '/') return 0;
    if (location.startsWith('/search')) return 1;
    if (location.startsWith('/messages')) return 2;
    if (location.startsWith('/map')) return 3;
    if (location.startsWith('/notifications')) return 4;
    if (location.startsWith('/profile')) return 4;
    return 0;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: child,
      bottomNavigationBar: NavigationBar(
        selectedIndex: _currentIndex(context),
        onDestinationSelected: (index) {
          switch (index) {
            case 0:
              context.go('/');
              break;
            case 1:
              context.go('/search');
              break;
            case 2:
              context.go('/messages');
              break;
            case 3:
              context.go('/map');
              break;
            case 4:
              context.go('/profile');
              break;
          }
        },
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.home_outlined),
            selectedIcon: Icon(Icons.home),
            label: 'Home',
          ),
          NavigationDestination(
            icon: Icon(Icons.search_outlined),
            selectedIcon: Icon(Icons.search),
            label: 'Search',
          ),
          NavigationDestination(
            icon: Icon(Icons.chat_bubble_outline),
            selectedIcon: Icon(Icons.chat_bubble),
            label: 'Messages',
          ),
          NavigationDestination(
            icon: Icon(Icons.map_outlined),
            selectedIcon: Icon(Icons.map),
            label: 'Map',
          ),
          NavigationDestination(
            icon: Icon(Icons.person_outlined),
            selectedIcon: Icon(Icons.person),
            label: 'Profile',
          ),
        ],
      ),
    );
  }
}
