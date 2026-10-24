'use client';

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useAuth } from '@/hooks/useAuth';
import { RegisterSchema, RegisterData } from '@/services/authService';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useState } from 'react';
import axios from 'axios';
import Spinner from '@/components/Spinner';

const FIELDS: { name: keyof RegisterData; label: string; type: string }[] = [
  { name: 'name', label: 'Name', type: 'text' },
  { name: 'last_name', label: 'Last name', type: 'text' },
  { name: 'username', label: 'Username', type: 'text' },
  { name: 'email', label: 'Email', type: 'email' },
  { name: 'phone', label: 'Phone', type: 'tel' },
  { name: 'password', label: 'Password', type: 'password' },
];

export default function RegisterPage() {
  const { register: registerUser } = useAuth();
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterData>({
    resolver: zodResolver(RegisterSchema),
  });

  const onSubmit = async (data: RegisterData) => {
    setLoading(true);
    setError(null);
    try {
      await registerUser(data);
      router.push(`/verify-email?email=${data.email}`);
    } catch (err) {
      setError(
        axios.isAxiosError(err) && err.response?.status === 409
          ? 'A user with this email or username already exists.'
          : 'Failed to register. Please try again.'
      );
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-50">
      <div className="w-full max-w-sm p-8 space-y-4 bg-white rounded-lg shadow-md">
        <h1 className="text-2xl font-bold text-center text-gray-800">Create an Account</h1>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          {error && <p className="text-sm text-red-500 text-center">{error}</p>}
          {FIELDS.map(({ name, label, type }) => (
            <div key={name}>
              <label
                htmlFor={name}
                className="block text-sm font-medium text-gray-600"
              >
                {label}
              </label>
              <input
                id={name}
                type={type}
                {...register(name)}
                className="block w-full px-3 py-2 mt-1 bg-gray-50 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
              />
              {errors[name] && (
                <p className="mt-1 text-sm text-red-500">{errors[name]?.message}</p>
              )}
            </div>
          ))}
          <button
            type="submit"
            disabled={loading}
            className="w-full flex justify-center px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:bg-blue-400"
          >
            {loading ? <Spinner /> : 'Register'}
          </button>
        </form>
        <p className="text-sm text-center text-gray-600">
          Already have an account?{' '}
          <Link href="/login" className="font-medium text-blue-600 hover:underline">
            Login
          </Link>
        </p>
      </div>
    </div>
  );
}