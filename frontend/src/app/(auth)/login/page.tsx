'use client';

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useAuth } from '@/hooks/useAuth';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useState } from 'react';
import Spinner from '@/components/Spinner';

const LoginSchema = z.object({
  user: z.string().email({ message: 'Invalid email address' }),
  password: z.string().min(6, { message: 'Password must be at least 6 characters' }),
});

type LoginData = z.infer<typeof LoginSchema>;

export default function LoginPage() {
  const { login } = useAuth();
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginData>({
    resolver: zodResolver(LoginSchema),
  });

  const onSubmit = async (data: LoginData) => {
    setLoading(true);
    setError(null);
    try {
      await login(data);
      router.push('/dashboard');
    } catch (err) {
      setError('Invalid email or password');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-whatsapp-deep-sea-green">
      <div className="w-full max-w-sm p-8 space-y-6 bg-white rounded-lg shadow-lg">
        <h1 className="text-3xl font-bold text-center text-whatsapp-surfie-green">
          Welcome Back
        </h1>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          {error && <p className="text-sm text-red-500 text-center">{error}</p>}
          <div>
            <label
              htmlFor="email"
              className="block text-sm font-medium text-gray-700"
            >
              Email Address
            </label>
            <input
              id="email"
              type="email"
              {...register('user')}
              className="block w-full px-4 py-3 mt-1 text-gray-900 bg-gray-100 border border-gray-300 rounded-md focus:outline-none focus:ring-whatsapp-surfie-green focus:border-whatsapp-surfie-green"
              placeholder="you@example.com"
            />
            {errors.user && (
              <p className="mt-2 text-sm text-red-600">{errors.user.message}</p>
            )}
          </div>
          <div>
            <label
              htmlFor="password"
              className="block text-sm font-medium text-gray-700"
            >
              Password
            </label>
            <input
              id="password"
              type="password"
              {...register('password')}
              className="block w-full px-4 py-3 mt-1 text-gray-900 bg-gray-100 border border-gray-300 rounded-md focus:outline-none focus:ring-whatsapp-surfie-green focus:border-whatsapp-surfie-green"
              placeholder="••••••••"
            />
            {errors.password && (
              <p className="mt-2 text-sm text-red-600">
                {errors.password.message}
              </p>
            )}
          </div>
          <button
            type="submit"
            disabled={loading}
            className="w-full flex justify-center py-3 px-4 text-sm font-medium text-white bg-whatsapp-surfie-green border border-transparent rounded-md shadow-sm hover:bg-whatsapp-deep-sea-green focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-whatsapp-surfie-green disabled:bg-opacity-50"
          >
            {loading ? <Spinner /> : 'Sign In'}
          </button>
        </form>
        <p className="text-sm text-center text-gray-600">
          Not a member?{' '}
          <Link href="/register" className="font-medium text-whatsapp-surfie-green hover:underline">
            Sign up
          </Link>
        </p>
      </div>
    </div>
  );
}