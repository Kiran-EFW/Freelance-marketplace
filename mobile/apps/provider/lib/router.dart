import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'main.dart';
import 'screens/dashboard/dashboard_screen.dart';
import 'screens/jobs/job_list_screen.dart';
import 'screens/jobs/job_detail_screen.dart';
import 'screens/earnings/earnings_screen.dart';
import 'screens/routes/route_list_screen.dart';
import 'screens/routes/route_detail_screen.dart';
import 'screens/auth/login_screen.dart';
import 'screens/profile/provider_profile_screen.dart';
import 'screens/kyc/kyc_screen.dart';
import 'screens/bank/bank_setup_screen.dart';

final routerProvider = Provider<GoRouter>((ref) {
  final authService = ref.watch(authServiceProvider);

  return GoRouter(
    initialLocation: '/',
    refreshListenable: authService,
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
          return _ProviderShell(child: child);
        },
        routes: [
          GoRoute(
            path: '/',
            name: 'dashboard',
            builder: (context, state) => const DashboardScreen(),
          ),
          GoRoute(
            path: '/jobs',
            name: 'jobs',
            builder: (context, state) => const JobListScreen(),
          ),
          GoRoute(
            path: '/earnings',
            name: 'earnings',
            builder: (context, state) => const EarningsScreen(),
          ),
          GoRoute(
            path: '/routes',
            name: 'routes',
            builder: (context, state) => const RouteListScreen(),
          ),
          GoRoute(
            path: '/profile',
            name: 'profile',
            builder: (context, state) => const ProviderProfileScreen(),
          ),
        ],
      ),

      // Detail routes (no bottom nav)
      GoRoute(
        path: '/job/:id',
        name: 'job-detail',
        builder: (context, state) {
          return ProviderJobDetailScreen(jobId: state.pathParameters['id']!);
        },
      ),
      GoRoute(
        path: '/route/:id',
        name: 'route-detail',
        builder: (context, state) {
          return RouteDetailScreen(routeId: state.pathParameters['id']!);
        },
      ),
      GoRoute(
        path: '/kyc',
        name: 'kyc',
        builder: (context, state) => const KycScreen(),
      ),
      GoRoute(
        path: '/bank-setup',
        name: 'bank-setup',
        builder: (context, state) => const BankSetupScreen(),
      ),

      // Auth
      GoRoute(
        path: '/auth/login',
        name: 'login',
        builder: (context, state) => const ProviderLoginScreen(),
      ),
    ],
  );
});

/// Shell widget providing bottom navigation for the provider app.
class _ProviderShell extends StatelessWidget {
  final Widget child;

  const _ProviderShell({required this.child});

  int _currentIndex(BuildContext context) {
    final location = GoRouterState.of(context).matchedLocation;
    if (location == '/') return 0;
    if (location.startsWith('/jobs')) return 1;
    if (location.startsWith('/earnings')) return 2;
    if (location.startsWith('/routes')) return 3;
    if (location.startsWith('/profile')) return 4;
    return 0;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: child,
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: _currentIndex(context),
        onTap: (index) {
          switch (index) {
            case 0:
              context.go('/');
              break;
            case 1:
              context.go('/jobs');
              break;
            case 2:
              context.go('/earnings');
              break;
            case 3:
              context.go('/routes');
              break;
            case 4:
              context.go('/profile');
              break;
          }
        },
        items: const [
          BottomNavigationBarItem(
            icon: Icon(Icons.dashboard_outlined),
            activeIcon: Icon(Icons.dashboard),
            label: 'Dashboard',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.work_outline),
            activeIcon: Icon(Icons.work),
            label: 'Jobs',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.account_balance_wallet_outlined),
            activeIcon: Icon(Icons.account_balance_wallet),
            label: 'Earnings',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.route_outlined),
            activeIcon: Icon(Icons.route),
            label: 'Routes',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.person_outlined),
            activeIcon: Icon(Icons.person),
            label: 'Profile',
          ),
        ],
      ),
    );
  }
}
