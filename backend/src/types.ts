const TYPES = {
  // Database
  DbPool: Symbol.for('DbPool'),

  // Repositories
  IMessageRepository: Symbol.for('IMessageRepository'),

  // Use Cases
  CreateMessageUseCase: Symbol.for('CreateMessageUseCase'),
};

export { TYPES };